package config

import (
	"k8sapi-pro/pkg/configmap"
	"k8sapi-pro/pkg/deployment"
	"k8sapi-pro/pkg/ingress"
	"k8sapi-pro/pkg/namespace"
	"k8sapi-pro/pkg/node"
	"k8sapi-pro/pkg/pod"
	"k8sapi-pro/pkg/rbac"
	"k8sapi-pro/pkg/secret"
	"k8sapi-pro/pkg/service"
)

type K8sService struct {
}

func NewK8sService() *K8sService {
	return &K8sService{}
}

func (this *K8sService) GetDeploySvc() *deployment.DeploySvc {
	return &deployment.DeploySvc{}
}

func (this *K8sService) GetPoSvc() *pod.PodSvc {
	return &pod.PodSvc{}
}

func (this *K8sService) GetPageHelper() *pod.PageHelper {
	return &pod.PageHelper{}
}

func (this *K8sService) GetIngressSvc() *ingress.IngressSvc {
	return &ingress.IngressSvc{}
}

func (this *K8sService) GetServiceSvc() *service.ServiceSvc {
	return &service.ServiceSvc{}
}

func (this *K8sService) GetCmSvc() *configmap.ConfigMapSvc {
	return &configmap.ConfigMapSvc{}
}

func (this *K8sService) GetSecretSvc() *secret.SecretSvc {
	return &secret.SecretSvc{}
}

func (this *K8sService) GetNodeSvc() *node.NodeSvc {
	return &node.NodeSvc{}
}

func (this *K8sService) GetNamespaceSvc() *namespace.NamespaceSvc {
	return &namespace.NamespaceSvc{}
}

func (this *K8sService) GetRoleSvc() *rbac.RoleSvc {
	return &rbac.RoleSvc{}
}

func (this *K8sService) GetSASvc() *rbac.SaSvc {
	return &rbac.SaSvc{}
}
