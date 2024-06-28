package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type RouteSpec struct {
	Version string `json:"version,omitempty"`
}
type Route struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RouteSpec `json:"spec,omitempty"`
}
type RouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Route `json:"items,omitempty"`
}

func init() {
	SchemeBuilder.Register(func(scheme *runtime.Scheme) error {
		gv := schema.GroupVersion{
			Group:   "extensions.octoboy.com",
			Version: "v1",
		}
		scheme.AddKnownTypes(gv, &Route{}, &RouteList{})
		metav1.AddToGroupVersion(scheme, gv)
		return nil
	}) //注册到scheme scheme：用于序列化api对象
}
