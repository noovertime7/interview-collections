package lib

import (
	"fmt"
	v1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strings"
)

func patchImagePullSecrets() []byte {
	str := `[
	  {
			"op" : "add" ,
			"path" : "/spec/imagePullSecrets" ,
			"value" : [{"name": "imagesecret"}]
		}
	]`
	//str := `[
	//     { "op": "add", "path": "/metadata/labels", "value": {"added-label": "yes"}}
	// ]`
	return []byte(str)
}

// AdmitPods 出处 https://github.com/kubernetes/kubernetes/blob/release-1.21/test/images/agnhost/webhook/pods.go
// only allow pods to pull images from specific registry.
func AdmitPods(ar v1.AdmissionReview) *v1.AdmissionResponse {
	klog.V(2).Info("admitting pods")
	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		err := fmt.Errorf("expect resource to be %s", podResource)
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	raw := ar.Request.Object.Raw
	pod := corev1.Pod{}
	deserializer := Codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(raw, nil, &pod); err != nil {
		klog.Error(err)
		return ToV1AdmissionResponse(err)
	}

	reviewResponse := v1.AdmissionResponse{}
	containers := pod.Spec.Containers
	//validate
	for _, container := range containers {
		if !strings.Contains(container.Image, "bestsign.tech") {
			reviewResponse.Allowed = false
			reviewResponse.Result = &metav1.Status{Code: 403,
				Message: "container's image must be from private hub."}
			return &reviewResponse
		}
	}

	// mutate
	if pod.Spec.ImagePullSecrets == nil {
		reviewResponse.Patch = patchImagePullSecrets()
		patchTypeJSONPatch := v1.PatchTypeJSONPatch
		reviewResponse.PatchType = &patchTypeJSONPatch
	}
	//reviewResponse.Patch = patchImagePullSecrets()
	//patchTypeJSONPatch := v1.PatchTypeJSONPatch
	//reviewResponse.PatchType = &patchTypeJSONPatch
	//放行
	reviewResponse.Allowed = true

	return &reviewResponse
}
