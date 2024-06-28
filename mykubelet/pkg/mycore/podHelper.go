package mycore

import (
	v1 "k8s.io/api/core/v1"
	runtimev1 "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/kubelet/container"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"time"
)

// 构造pod状态ready
func GetPodStatusReady(po *v1.Pod) *container.PodStatus {
	containerStatus := make([]*container.Status, len(po.Spec.Containers))
	for i, container_item := range po.Spec.Containers {
		containerStatus[i] = &container.Status{
			Name:  container_item.Name,
			Image: container_item.Image,
			State: container.ContainerStateRunning,
		}
	}

	return &container.PodStatus{
		ID:        po.UID,
		Name:      po.Name,
		Namespace: po.Namespace,
		SandboxStatuses: []*runtimev1.PodSandboxStatus{
			{
				Id:    string(po.UID),
				State: runtimev1.PodSandboxState_SANDBOX_READY,
			},
		},
		ContainerStatuses: containerStatus,
	}
}

// 构造pod状态completed
func GetPodStatusCompleted(po *v1.Pod) *container.PodStatus {
	containerStatus := make([]*container.Status, len(po.Spec.Containers))
	for i, container_item := range po.Spec.Containers {
		containerStatus[i] = &container.Status{
			Name:       container_item.Name,
			Image:      container_item.Image,
			State:      container.ContainerStateExited,
			ExitCode:   0,
			Reason:     "Completed",
			FinishedAt: time.Now(),
		}
	}

	return &container.PodStatus{
		ID:        po.UID,
		Name:      po.Name,
		Namespace: po.Namespace,
		SandboxStatuses: []*runtimev1.PodSandboxStatus{
			{
				Id:    string(po.UID),
				State: runtimev1.PodSandboxState_SANDBOX_NOTREADY,
			},
		},
		ContainerStatuses: containerStatus,
	}
}

func GetPodStatusExited(po *v1.Pod, containerName string, exitCode int) *container.PodStatus {
	containerStatus := make([]*container.Status, len(po.Spec.Containers))
	for i, container_item := range po.Spec.Containers {
		if container_item.Name != containerName {
			continue
		}
		reason := "Completed"
		if exitCode != 0 {
			reason = "Error"
		}
		containerStatus[i] = &container.Status{
			Name:       container_item.Name,
			Image:      container_item.Image,
			State:      container.ContainerStateExited,
			ExitCode:   exitCode,
			Reason:     reason,
			FinishedAt: time.Now(),
		}
	}
	state := runtimev1.PodSandboxState_SANDBOX_NOTREADY
	if len(po.Spec.Containers) > 1 {
		state = runtimev1.PodSandboxState_SANDBOX_READY
	}
	return &container.PodStatus{
		ID:        po.UID,
		Name:      po.Name,
		Namespace: po.Namespace,
		SandboxStatuses: []*runtimev1.PodSandboxStatus{
			{
				Id:    string(po.UID),
				State: state,
			},
		},
		ContainerStatuses: containerStatus,
	}
}

func HandlerAddPod(pods []*v1.Pod, pc *PodCache, f CallBackFunc) {
	for _, pod := range pods {
		pc.PodManager.AddPod(pod) //将pod存入map缓存
		// 这一步相当于dispatchWork，调用后会开携程不断监测pod状态，一旦有变化就调用syncPod同步（managePodLoop）
		pc.PodWorker.UpdatePod(UpdatePodOptions{
			StartTime:  pc.Clock.Now(),
			Pod:        pod,
			MirrorPod:  nil,
			UpdateType: kubetypes.SyncPodCreate,
		})
		if f != nil {
			c := &CallBackContent{
				Pod:      pod,
				recorder: pc.PodWorker.(*podWorkers).recorder,
				PodCache: pc,
			}
			err := f(c)
			if err != nil {
				klog.Error(err)
			}
		}
	}
}

func HandlerUpdatePod(pods []*v1.Pod, pc *PodCache, f CallBackFunc) {
	for _, pod := range pods {
		pc.PodManager.UpdatePod(pod)
		pc.PodWorker.UpdatePod(UpdatePodOptions{
			StartTime:  pc.Clock.Now(),
			Pod:        pod,
			MirrorPod:  nil,
			UpdateType: kubetypes.SyncPodUpdate,
		})
		if f != nil {
			c := &CallBackContent{
				Pod:      pod,
				recorder: pc.PodWorker.(*podWorkers).recorder,
				PodCache: pc,
			}
			err := f(c)
			if err != nil {
				klog.Error(err)
			}
		}
	}
}

func HandlerRemovePod(pods []*v1.Pod, pc *PodCache, f CallBackFunc) {
	for _, pod := range pods {
		pc.PodManager.DeletePod(pod)
		pc.PodWorker.UpdatePod(UpdatePodOptions{
			Pod:        pod,
			UpdateType: kubetypes.SyncPodKill,
		})
		if f != nil {
			c := &CallBackContent{
				Pod:      pod,
				recorder: pc.PodWorker.(*podWorkers).recorder,
				PodCache: pc,
			}
			err := f(c)
			if err != nil {
				klog.Error(err)
			}
		}
	}
}
