package helper

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
)

func GetResourceYaml(cf *genericclioptions.ConfigFlags, typ, name string) ([]byte, error) {
	ns, _, _ := cf.ToRawKubeConfigLoader().Namespace()
	obj, err := resource.NewBuilder(cf).
		DefaultNamespace().
		NamespaceParam(ns).
		SingleResourceType().
		ResourceNames(typ, name).
		Unstructured().
		Do().
		Object()
	if err != nil {
		return nil, fmt.Errorf("get object: %w", err)
	}

	metaObj, err := meta.Accessor(obj)
	if err != nil {
		return nil, fmt.Errorf("access object: %w", err)
	} else {
		metaObj.SetManagedFields(nil)
	}

	b, err := yaml.Marshal(metaObj)
	if err != nil {
		return nil, fmt.Errorf("marshal object: %w", err)
	}

	return b, nil
}
