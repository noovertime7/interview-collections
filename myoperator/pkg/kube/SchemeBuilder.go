package kube

import "k8s.io/apimachinery/pkg/runtime"

var SchemeBuilder = &Builder{}

type Builder struct {
	runtime.SchemeBuilder
}

func (this *Builder) AddSceme(scheme *runtime.Scheme) error {
	return this.AddToScheme(scheme)
}
