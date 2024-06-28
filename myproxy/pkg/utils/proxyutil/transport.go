package proxyutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/http"
	"regexp"
	"strconv"
)

// 也同样还是嵌套
// http.DefaultClient.Do()其实就是死循环执行了transport的RoundTrip方法。然后做了一些额外处理，比如返回302，直到得到200跳出循环
type MyTransport struct {
	http.RoundTripper
}

// 这里可以拦截响应，对响应内容进行处理
func (mtp *MyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	//fmt.Println("深层拦截：" + req.RequestURI)
	rsp, err := mtp.RoundTripper.RoundTrip(req)

	if err != nil {
		fmt.Println("请求失败：" + err.Error())
		return nil, err
	}

	//针对/openapi/v2做处理（apply会先发这个请求）
	if regexp.MustCompile(OpenApiPattern).MatchString(req.RequestURI) {
		return rsp, nil
	}

	defer rsp.Body.Close()

	gr := rsp.Body

	b, err := ioutil.ReadAll(gr) //读取后，rsp body内容将就空了。需要重新赋值回去
	//fmt.Println("响应内容：" + string(b))
	if err != nil {
		return nil, err
	}
	//序列化对象
	obj := unstructured.Unstructured{}
	if err = obj.UnmarshalJSON(b); err == nil {
		addCustomColumn(&obj, req)

		//如果获取api-resource，加入自定义资源
		handlerMyResource(&obj, req)
		b, err = obj.MarshalJSON()
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println(string(b)) //这里遇到了返回service unavailable导致解析失败，原因未知
		return nil, err
	}
	rsp.Body = ioutil.NopCloser(bytes.NewReader(b)) //tips 给body赋值的方法
	//fmt.Println("响应对象类型:" + obj.GetKind())
	rsp.Header.Set("Content-Length", strconv.Itoa(len(b))) //这个很重要，如果修改了body内容需要重新计算长度
	return rsp, nil
}

func WrapperTransport(tp http.RoundTripper) *MyTransport {
	return &MyTransport{
		tp,
	}
}
