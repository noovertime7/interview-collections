package config

import (
	"k8sapi-pro/pkg/namespace"
	"k8sapi-pro/src/caches"
)

type K8sHandler struct {
}

func NewK8sHandler() *K8sHandler {
	return &K8sHandler{}
}

func (this *K8sHandler) ResourceHandler() *caches.ResourceHandler {
	return &caches.ResourceHandler{}
}

func (this *K8sHandler) NsHandler() *namespace.NsHandler {
	return &namespace.NsHandler{}
}
