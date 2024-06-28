package utils

import (
	"github.com/shenyisyn/goft-gin/goft"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
)

func Convert(obj []byte) interface{} {
	o := &unstructured.Unstructured{}
	goft.Error(json.Unmarshal(obj, o))
	kind := o.GetKind()
	switch kind {
	case "Deployment":
		deploy := &v1.Deployment{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, deploy))
		return deploy
	case "Pod":
		pod := &corev1.Pod{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, pod))
		return pod
	case "Namespace":
		ns := &corev1.Namespace{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, ns))
		return ns
	case "Ingress":
		ingress := &netv1.Ingress{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, ingress))
		return ingress
	case "Service":
		service := &corev1.Service{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, service))
		return service
	case "ConfigMap":
		cm := &corev1.ConfigMap{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, cm))
		return cm
	case "Secret":
		secret := &corev1.Secret{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, secret))
		return secret
	case "Node":
		node := &corev1.Node{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, node))
		return node
	case "Role":
		role := &rbacv1.Role{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, role))
		return role
	case "ClusterRole":
		crole := &rbacv1.ClusterRole{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, crole))
		return crole
	case "RoleBinding":
		roleb := &rbacv1.RoleBinding{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, roleb))
		return roleb
	case "ClusterRoleBinding":
		croleb := &rbacv1.ClusterRoleBinding{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, croleb))
		return croleb
	case "ServiceAccount":
		sa := &corev1.ServiceAccount{}
		goft.Error(runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, sa))
		return sa
	default:
		return nil
	}
}
