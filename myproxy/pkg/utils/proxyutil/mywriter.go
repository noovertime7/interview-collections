package proxyutil

import "net/http"

func WrapWriter() *MyWriter {
	return &MyWriter{content: make([][]byte, 0)}
}

type MyWriter struct {
	http.ResponseWriter
	headerCode int
	content    [][]byte //保存内容
}

func (w *MyWriter) Header() http.Header {
	return map[string][]string{
		"Content-type": []string{"application/json"},
	}

}
func (w *MyWriter) Write(b []byte) (int, error) {
	w.content = append(w.content, b)
	return len(b), nil
}
func (w *MyWriter) WriteHeader(h int) {
	w.headerCode = h
	return
}
