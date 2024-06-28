package deployment

import "k8sapi-pro/pkg/pod"

type Deployment struct {
	Name        string
	NameSpace   string
	Replicas    [3]int32 //3个值，分别是总副本数，可用副本数 ，不可用副本数
	Images      string
	CreateTime  string
	Pods        []*pod.Pod
	IsCompleted bool
	Message     string
}
