package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"myhook/lib"
	"net/http"
)

func main() {
	http.HandleFunc("/pods", func(writer http.ResponseWriter, request *http.Request) {
		var body []byte
		if request.Body != nil {
			if data, err := ioutil.ReadAll(request.Body); err == nil {
				body = data
			}
		}
		fmt.Println(request.Header)
		//第二步
		reqAdmissionReview := v1.AdmissionReview{} //请求
		resAdmissionReview := v1.AdmissionReview{  //响应 ---完整的对象在 https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/#response
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
		}
		//第三步，把body decode成对象
		deserializer := lib.Codecs.UniversalDeserializer()
		if _, _, err := deserializer.Decode(body, nil, &reqAdmissionReview); err != nil {
			resAdmissionReview.Response = lib.ToV1AdmissionResponse(err)
		} else {
			resAdmissionReview.Response = lib.AdmitPods(reqAdmissionReview)
		}

		resAdmissionReview.Response.UID = reqAdmissionReview.Request.UID
		marshal, _ := json.Marshal(resAdmissionReview)
		_, err := writer.Write(marshal)
		if err != nil {
			return
		}
	})

	tlsConfig := lib.Config{
		CertFile: "/etc/webhook/certs/tls.crt",
		KeyFile:  "/etc/webhook/certs/tls.key",
	}

	server := http.Server{
		Addr:      ":443",
		TLSConfig: lib.ConfigTLS(tlsConfig),
	}

	err := server.ListenAndServeTLS("", "")
	if err != nil {
		return
	}
}
