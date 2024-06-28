package mycore

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
)

func getPhase(spec *v1.PodSpec, info []v1.ContainerStatus) v1.PodPhase {
	pendingInitialization := 0
	failedInitialization := 0
	for _, container := range spec.InitContainers {
		containerStatus, ok := podutil.GetContainerStatus(info, container.Name)
		if !ok {
			pendingInitialization++
			continue
		}

		switch {
		case containerStatus.State.Running != nil:
			pendingInitialization++
		case containerStatus.State.Terminated != nil:
			if containerStatus.State.Terminated.ExitCode != 0 {
				failedInitialization++
			}
		case containerStatus.State.Waiting != nil:
			if containerStatus.LastTerminationState.Terminated != nil {
				if containerStatus.LastTerminationState.Terminated.ExitCode != 0 {
					failedInitialization++
				}
			} else {
				pendingInitialization++
			}
		default:
			pendingInitialization++
		}
	}

	unknown := 0
	running := 0
	waiting := 0
	stopped := 0
	succeeded := 0
	for _, container := range spec.Containers {
		containerStatus, ok := podutil.GetContainerStatus(info, container.Name)
		if !ok {
			unknown++
			continue
		}

		switch {
		case containerStatus.State.Running != nil:
			running++
		case containerStatus.State.Terminated != nil:
			stopped++
			if containerStatus.State.Terminated.ExitCode == 0 {
				succeeded++
			}
		case containerStatus.State.Waiting != nil:
			if containerStatus.LastTerminationState.Terminated != nil {
				stopped++
			} else {
				waiting++
			}
		default:
			unknown++
		}
	}

	if failedInitialization > 0 && spec.RestartPolicy == v1.RestartPolicyNever {
		return v1.PodFailed
	}

	switch {
	case pendingInitialization > 0:
		fallthrough
	case waiting > 0:
		klog.V(5).InfoS("Pod waiting > 0, pending")
		// One or more containers has not been started
		return v1.PodPending
	case running > 0 && unknown == 0:
		// All containers have been started, and at least
		// one container is running
		return v1.PodRunning
	case running == 0 && stopped > 0 && unknown == 0:
		// All containers are terminated
		if spec.RestartPolicy == v1.RestartPolicyAlways {
			// All containers are in the process of restarting
			return v1.PodRunning
		}
		if stopped == succeeded {
			// RestartPolicy is not Always, and all
			// containers are terminated in success
			return v1.PodSucceeded
		}
		if spec.RestartPolicy == v1.RestartPolicyNever {
			// RestartPolicy is Never, and all containers are
			// terminated with at least one in failure
			return v1.PodFailed
		}
		// RestartPolicy is OnFailure, and at least one in failure
		// and in the process of restarting
		return v1.PodRunning
	default:
		klog.V(5).InfoS("Pod default case, pending")
		return v1.PodPending
	}
}
