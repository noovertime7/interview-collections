package config

import "k8sapi-pro/pkg/namespace"

type K8sMap struct {
}

func NewK8sMap() *K8sMap {
	return &K8sMap{}
}

func (this *K8sMap) GetNsMap() *namespace.NsMap {
	return &namespace.NsMap{}
}
