package mycore

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	v1qos "k8s.io/kubernetes/pkg/apis/core/v1/helper/qos"
	"k8s.io/kubernetes/pkg/kubelet"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	"k8s.io/kubernetes/pkg/kubelet/prober"
	"k8s.io/kubernetes/pkg/kubelet/prober/results"
	"k8s.io/kubernetes/pkg/kubelet/status"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"sort"
)

// 实际的 创建container 和 把状态同步到apiServer的地方
func SyncPodFn(ctx context.Context, updateType kubetypes.SyncPodType, pod *v1.Pod, mirrorPod *v1.Pod, podStatus *kubecontainer.PodStatus) (bool, error) {
	fmt.Println("临时的syncPodFn")
	return true, nil
}
func SyncTerminatingPodFn(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus, runningPod *kubecontainer.Pod, gracePeriod *int64, podStatusFn func(*v1.PodStatus)) error {
	fmt.Println("临时的syncTerminatingPodFn")
	return nil
}
func SyncTerminatedPodFn(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus) error {
	fmt.Println("临时的syncTerminatedPodFn")
	return nil
}

type PodFn struct {
	kubeClient    clientset.Interface
	statusManager status.Manager
	reasonCache   *kubelet.ReasonCache
	probeManager  prober.Manager
	recorder      record.EventRecorder
}

func NewPodFn(kubeClient clientset.Interface, statusManager status.Manager, recorder record.EventRecorder) *PodFn {
	lm, rm, sm := results.NewManager(), results.NewManager(), results.NewManager()
	return &PodFn{
		kubeClient:    kubeClient,
		statusManager: statusManager,
		probeManager:  prober.NewManager(statusManager, lm, rm, sm, nil, recorder),
		recorder:      recorder,
		reasonCache:   kubelet.NewReasonCache(),
	}
}

const (
	PodInitializing   = "PodInitializing"
	ContainerCreating = "ContainerCreating"
)

func (kl *PodFn) convertToAPIContainerStatuses(pod *v1.Pod, podStatus *kubecontainer.PodStatus, previousStatus []v1.ContainerStatus, containers []v1.Container, hasInitContainers, isInitContainer bool) []v1.ContainerStatus {
	convertContainerStatus := func(cs *kubecontainer.Status, oldStatus *v1.ContainerStatus) *v1.ContainerStatus {
		cid := cs.ID.String()
		status := &v1.ContainerStatus{
			Name:         cs.Name,
			RestartCount: int32(cs.RestartCount),
			Image:        cs.Image,
			ImageID:      cs.ImageID,
			ContainerID:  cid,
		}
		switch {
		case cs.State == kubecontainer.ContainerStateRunning:
			status.State.Running = &v1.ContainerStateRunning{StartedAt: metav1.NewTime(cs.StartedAt)}
		case cs.State == kubecontainer.ContainerStateCreated:
			// containers that are created but not running are "waiting to be running"
			status.State.Waiting = &v1.ContainerStateWaiting{}
		case cs.State == kubecontainer.ContainerStateExited:
			status.State.Terminated = &v1.ContainerStateTerminated{
				ExitCode:    int32(cs.ExitCode),
				Reason:      cs.Reason,
				Message:     cs.Message,
				StartedAt:   metav1.NewTime(cs.StartedAt),
				FinishedAt:  metav1.NewTime(cs.FinishedAt),
				ContainerID: cid,
			}

		case cs.State == kubecontainer.ContainerStateUnknown &&
			oldStatus != nil && // we have an old status
			oldStatus.State.Running != nil: // our previous status was running
			// if this happens, then we know that this container was previously running and isn't anymore (assuming the CRI isn't failing to return running containers).
			// you can imagine this happening in cases where a container failed and the kubelet didn't ask about it in time to see the result.
			// in this case, the container should not to into waiting state immediately because that can make cases like runonce pods actually run
			// twice. "container never ran" is different than "container ran and failed".  This is handled differently in the kubelet
			// and it is handled differently in higher order logic like crashloop detection and handling
			status.State.Terminated = &v1.ContainerStateTerminated{
				Reason:   "ContainerStatusUnknown",
				Message:  "The container could not be located when the pod was terminated",
				ExitCode: 137, // this code indicates an error
			}
			// the restart count normally comes from the CRI (see near the top of this method), but since this is being added explicitly
			// for the case where the CRI did not return a status, we need to manually increment the restart count to be accurate.
			status.RestartCount = oldStatus.RestartCount + 1

		default:
			// this collapses any unknown state to container waiting.  If any container is waiting, then the pod status moves to pending even if it is running.
			// if I'm reading this correctly, then any failure to read status on any container results in the entire pod going pending even if the containers
			// are actually running.
			// see https://github.com/kubernetes/kubernetes/blob/5d1b3e26af73dde33ecb6a3e69fb5876ceab192f/pkg/kubelet/kuberuntime/kuberuntime_container.go#L497 to
			// https://github.com/kubernetes/kubernetes/blob/8976e3620f8963e72084971d9d4decbd026bf49f/pkg/kubelet/kuberuntime/helpers.go#L58-L71
			// and interpreted here https://github.com/kubernetes/kubernetes/blob/b27e78f590a0d43e4a23ca3b2bf1739ca4c6e109/pkg/kubelet/kubelet_pods.go#L1434-L1439
			status.State.Waiting = &v1.ContainerStateWaiting{}
		}
		return status
	}

	// Fetch old containers statuses from old pod status.
	oldStatuses := make(map[string]v1.ContainerStatus, len(containers))
	for _, status := range previousStatus {
		oldStatuses[status.Name] = status
	}

	// Set all container statuses to default waiting state
	statuses := make(map[string]*v1.ContainerStatus, len(containers))
	defaultWaitingState := v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: ContainerCreating}}
	if hasInitContainers {
		defaultWaitingState = v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: PodInitializing}}
	}

	for _, container := range containers {
		status := &v1.ContainerStatus{
			Name:  container.Name,
			Image: container.Image,
			State: defaultWaitingState,
		}
		oldStatus, found := oldStatuses[container.Name]
		if found {
			if oldStatus.State.Terminated != nil {
				status = &oldStatus
			} else {
				// Apply some values from the old statuses as the default values.
				status.RestartCount = oldStatus.RestartCount
				status.LastTerminationState = oldStatus.LastTerminationState
			}
		}
		statuses[container.Name] = status
	}

	for _, container := range containers {
		found := false
		for _, cStatus := range podStatus.ContainerStatuses {
			if container.Name == cStatus.Name {
				found = true
				break
			}
		}
		if found {
			continue
		}
		// if no container is found, then assuming it should be waiting seems plausible, but the status code requires
		// that a previous termination be present.  If we're offline long enough or something removed the container, then
		// the previous termination may not be present.  This next code block ensures that if the container was previously running
		// then when that container status disappears, we can infer that it terminated even if we don't know the status code.
		// By setting the lasttermination state we are able to leave the container status waiting and present more accurate
		// data via the API.

		oldStatus, ok := oldStatuses[container.Name]
		if !ok {
			continue
		}
		if oldStatus.State.Terminated != nil {
			// if the old container status was terminated, the lasttermination status is correct
			continue
		}
		if oldStatus.State.Running == nil {
			// if the old container status isn't running, then waiting is an appropriate status and we have nothing to do
			continue
		}

		// If we're here, we know the pod was previously running, but doesn't have a terminated status. We will check now to
		// see if it's in a pending state.
		status := statuses[container.Name]
		// If the status we're about to write indicates the default, the Waiting status will force this pod back into Pending.
		// That isn't true, we know the pod was previously running.
		isDefaultWaitingStatus := status.State.Waiting != nil && status.State.Waiting.Reason == ContainerCreating
		if hasInitContainers {
			isDefaultWaitingStatus = status.State.Waiting != nil && status.State.Waiting.Reason == PodInitializing
		}
		if !isDefaultWaitingStatus {
			// the status was written, don't override
			continue
		}
		if status.LastTerminationState.Terminated != nil {
			// if we already have a termination state, nothing to do
			continue
		}

		// setting this value ensures that we show as stopped here, not as waiting:
		// https://github.com/kubernetes/kubernetes/blob/90c9f7b3e198e82a756a68ffeac978a00d606e55/pkg/kubelet/kubelet_pods.go#L1440-L1445
		// This prevents the pod from becoming pending
		status.LastTerminationState.Terminated = &v1.ContainerStateTerminated{
			Reason:   "ContainerStatusUnknown",
			Message:  "The container could not be located when the pod was deleted.  The container used to be Running",
			ExitCode: 137,
		}

		// If the pod was not deleted, then it's been restarted. Increment restart count.
		if pod.DeletionTimestamp == nil {
			status.RestartCount += 1
		}

		statuses[container.Name] = status
	}

	// Copy the slice before sorting it
	containerStatusesCopy := make([]*kubecontainer.Status, len(podStatus.ContainerStatuses))
	copy(containerStatusesCopy, podStatus.ContainerStatuses)

	// Make the latest container status comes first.
	sort.Sort(sort.Reverse(kubecontainer.SortContainerStatusesByCreationTime(containerStatusesCopy)))
	// Set container statuses according to the statuses seen in pod status
	containerSeen := map[string]int{}
	for _, cStatus := range containerStatusesCopy {
		cName := cStatus.Name
		if _, ok := statuses[cName]; !ok {
			// This would also ignore the infra container.
			continue
		}
		if containerSeen[cName] >= 2 {
			continue
		}
		var oldStatusPtr *v1.ContainerStatus
		if oldStatus, ok := oldStatuses[cName]; ok {
			oldStatusPtr = &oldStatus
		}
		status := convertContainerStatus(cStatus, oldStatusPtr)
		if containerSeen[cName] == 0 {
			statuses[cName] = status
		} else {
			statuses[cName].LastTerminationState = status.State
		}
		containerSeen[cName] = containerSeen[cName] + 1
	}

	// Handle the containers failed to be started, which should be in Waiting state.
	for _, container := range containers {
		if isInitContainer {
			// If the init container is terminated with exit code 0, it won't be restarted.
			// TODO(random-liu): Handle this in a cleaner way.
			s := podStatus.FindContainerStatusByName(container.Name)
			if s != nil && s.State == kubecontainer.ContainerStateExited && s.ExitCode == 0 {
				continue
			}
		}
		// If a container should be restarted in next syncpod, it is *Waiting*.
		if !kubecontainer.ShouldContainerBeRestarted(&container, pod, podStatus) {
			continue
		}
		status := statuses[container.Name]
		reason, ok := kl.reasonCache.Get(pod.UID, container.Name)
		if !ok {
			// In fact, we could also apply Waiting state here, but it is less informative,
			// and the container will be restarted soon, so we prefer the original state here.
			// Note that with the current implementation of ShouldContainerBeRestarted the original state here
			// could be:
			//   * Waiting: There is no associated historical container and start failure reason record.
			//   * Terminated: The container is terminated.
			continue
		}
		if status.State.Terminated != nil {
			status.LastTerminationState = status.State
		}
		status.State = v1.ContainerState{
			Waiting: &v1.ContainerStateWaiting{
				Reason:  reason.Err.Error(),
				Message: reason.Message,
			},
		}
		statuses[container.Name] = status
	}

	// Sort the container statuses since clients of this interface expect the list
	// of containers in a pod has a deterministic order.
	if isInitContainer {
		return kubetypes.SortStatusesOfInitContainers(pod, statuses)
	}
	containerStatuses := make([]v1.ContainerStatus, 0, len(statuses))
	for _, status := range statuses {
		containerStatuses = append(containerStatuses, *status)
	}

	sort.Sort(kubetypes.SortedContainerStatuses(containerStatuses))
	return containerStatuses
}

func (kl *PodFn) convertStatusToAPIStatus(pod *v1.Pod, podStatus *kubecontainer.PodStatus, oldPodStatus v1.PodStatus) *v1.PodStatus {
	var apiPodStatus v1.PodStatus

	// copy pod status IPs to avoid race conditions with PodStatus #102806
	podIPs := make([]string, len(podStatus.IPs))
	for j, ip := range podStatus.IPs {
		podIPs[j] = ip
	}

	// set status for Pods created on versions of kube older than 1.6
	apiPodStatus.QOSClass = v1qos.GetPodQOS(pod)

	apiPodStatus.ContainerStatuses = kl.convertToAPIContainerStatuses(
		pod, podStatus,
		oldPodStatus.ContainerStatuses,
		pod.Spec.Containers,
		len(pod.Spec.InitContainers) > 0,
		false,
	)
	apiPodStatus.InitContainerStatuses = kl.convertToAPIContainerStatuses(
		pod, podStatus,
		oldPodStatus.InitContainerStatuses,
		pod.Spec.InitContainers,
		len(pod.Spec.InitContainers) > 0,
		true,
	)
	var ecSpecs []v1.Container
	for i := range pod.Spec.EphemeralContainers {
		ecSpecs = append(ecSpecs, v1.Container(pod.Spec.EphemeralContainers[i].EphemeralContainerCommon))
	}

	// #80875: By now we've iterated podStatus 3 times. We could refactor this to make a single
	// pass through podStatus.ContainerStatuses
	apiPodStatus.EphemeralContainerStatuses = kl.convertToAPIContainerStatuses(
		pod, podStatus,
		oldPodStatus.EphemeralContainerStatuses,
		ecSpecs,
		len(pod.Spec.InitContainers) > 0,
		false,
	)

	return &apiPodStatus
}

func (kl *PodFn) generateAPIPodStatus(pod *v1.Pod, podStatus *kubecontainer.PodStatus) v1.PodStatus {
	klog.V(3).InfoS("Generating pod status", "pod", klog.KObj(pod))
	// use the previous pod status, or the api status, as the basis for this pod
	oldPodStatus, found := kl.statusManager.GetPodStatus(pod.UID)
	if !found {
		oldPodStatus = pod.Status
	}
	s := kl.convertStatusToAPIStatus(pod, podStatus, oldPodStatus)
	// calculate the next phase and preserve reason
	allStatus := append(append([]v1.ContainerStatus{}, s.ContainerStatuses...), s.InitContainerStatuses...)
	s.Phase = getPhase(&pod.Spec, allStatus)
	klog.V(4).InfoS("Got phase for pod", "pod", klog.KObj(pod), "oldPhase", oldPodStatus.Phase, "phase", s.Phase)

	// Perform a three-way merge between the statuses from the status manager,
	// runtime, and generated status to ensure terminal status is correctly set.
	if s.Phase != v1.PodFailed && s.Phase != v1.PodSucceeded {
		switch {
		case oldPodStatus.Phase == v1.PodFailed || oldPodStatus.Phase == v1.PodSucceeded:
			klog.V(4).InfoS("Status manager phase was terminal, updating phase to match", "pod", klog.KObj(pod), "phase", oldPodStatus.Phase)
			s.Phase = oldPodStatus.Phase
		case pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded:
			klog.V(4).InfoS("API phase was terminal, updating phase to match", "pod", klog.KObj(pod), "phase", pod.Status.Phase)
			s.Phase = pod.Status.Phase
		}
	}

	// ensure the probe managers have up to date status for containers
	kl.probeManager.UpdatePodStatus(pod.UID, s)

	if s.Phase == oldPodStatus.Phase {
		// preserve the reason and message which is associated with the phase
		s.Reason = oldPodStatus.Reason
		s.Message = oldPodStatus.Message
		if len(s.Reason) == 0 {
			s.Reason = pod.Status.Reason
		}
		if len(s.Message) == 0 {
			s.Message = pod.Status.Message
		}
	}

	// check if an internal module has requested the pod is evicted and override the reason and message

	// pods are not allowed to transition out of terminal phases
	if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
		// API server shows terminal phase; transitions are not allowed
		if s.Phase != pod.Status.Phase {
			klog.ErrorS(nil, "Pod attempted illegal phase transition", "pod", klog.KObj(pod), "originalStatusPhase", pod.Status.Phase, "apiStatusPhase", s.Phase, "apiStatus", s)
			// Force back to phase from the API server
			s.Phase = pod.Status.Phase
		}
	}

	// preserve all conditions not owned by the kubelet
	s.Conditions = make([]v1.PodCondition, 0, len(pod.Status.Conditions)+1)
	for _, c := range pod.Status.Conditions {
		if !kubetypes.PodConditionByKubelet(c.Type) {
			s.Conditions = append(s.Conditions, c)
		}
	}
	// set all Kubelet-owned conditions
	//注释了 默认是关闭的
	//if utilfeature.DefaultFeatureGate.Enabled(features.PodHasNetworkCondition) {
	//	s.Conditions = append(s.Conditions, status.GeneratePodHasNetworkCondition(pod, podStatus))
	//}
	s.Conditions = append(s.Conditions, status.GeneratePodInitializedCondition(&pod.Spec, s.InitContainerStatuses, s.Phase))
	s.Conditions = append(s.Conditions, status.GeneratePodReadyCondition(&pod.Spec, s.Conditions, s.ContainerStatuses, s.Phase))
	s.Conditions = append(s.Conditions, status.GenerateContainersReadyCondition(&pod.Spec, s.ContainerStatuses, s.Phase))
	s.Conditions = append(s.Conditions, v1.PodCondition{
		Type:   v1.PodScheduled,
		Status: v1.ConditionTrue,
	})

	return *s
}

//以上三个函数是为了 把cri的PodStatus转化成供api调用的PodStatus
//tips 并且包括了 调用probeManager去修改Conditions

// 这里是调用apiServer同步PodStatus的地方
func (kl *PodFn) SyncPodFn(ctx context.Context, updateType kubetypes.SyncPodType, pod *v1.Pod, mirrorPod *v1.Pod, podStatus *kubecontainer.PodStatus) (bool, error) {

	pod.Status = kl.generateAPIPodStatus(pod, podStatus)
	//这里 整合了所有对pod的api操作，包括同步状态以及是否能删除等等
	//tips 删除的逻辑是这样：1,kubectl delete访问apiServer打上deletionTimestamp
	// tips 2, kubelet同步到本地，判断能够删除后通过statusManager再同步给apiServer
	kl.statusManager.SetPodStatus(pod, pod.Status)

	return true, nil
}

// 把completed或者terminating同步到apiServer
func (kl *PodFn) SyncTerminatingPodFn(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus, runningPod *kubecontainer.Pod, gracePeriod *int64, podStatusFn func(*v1.PodStatus)) error {
	pod.Status = kl.generateAPIPodStatus(pod, podStatus)
	kl.statusManager.SetPodStatus(pod, pod.Status)
	return nil
}

func (kl *PodFn) SyncTerminatedPodFn(ctx context.Context, pod *v1.Pod, podStatus *kubecontainer.PodStatus) error {
	return nil
}
