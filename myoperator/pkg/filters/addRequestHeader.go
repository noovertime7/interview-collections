package filters

import (
	"github.com/valyala/fasthttp"
	"strings"
)

const request_header_annotation_key = "octoboy.ingress.kubernetes.io/add-request-header"

func init() {
	//tips nil 没办法强制转换成RewriteFilter，但是nil可以转换成*RewriteFilter，代表空指针
	registerFilter(request_header_annotation_key, (*RequestHeaderFilter)(nil))
}

type RequestHeaderFilter struct {
	PathReg string
	Target  string
}

func (r *RequestHeaderFilter) SetValue(value ...string) {
	r.Target = value[0]
}

func (r *RequestHeaderFilter) SetPathReg(value ...string) {
}

func (r *RequestHeaderFilter) Do(ctx *fasthttp.RequestCtx) {
	kvs := strings.Split(r.Target, ";")
	for _, kv := range kvs {
		set := strings.Split(kv, "=")
		ctx.Request.Header.Add(set[0], set[1])
	}

}

var _ ProxyFilter = &RequestHeaderFilter{}
