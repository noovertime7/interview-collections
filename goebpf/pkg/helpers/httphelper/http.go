package httphelper

import (
	"bufio"
	"bytes"
	"net/http"
)

// 根据报文判断是否是http请求
func IsHttpRequest(payload []byte) (*http.Request, bool) {
	reader := bufio.NewReader(bytes.NewBuffer(payload))
	req, err := http.ReadRequest(reader)
	if err != nil {
		return nil, false
	}
	return req, true
}

// 根据报文判断是否是http响应
func IsHttpResponse(payload []byte) (*http.Response, bool) {
	reader := bufio.NewReader(bytes.NewBuffer(payload))
	resp, err := http.ReadResponse(reader, nil)
	if err != nil {
		return nil, false
	}
	return resp, true
}
