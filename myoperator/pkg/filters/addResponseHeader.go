package filters

import (
	"github.com/valyala/fasthttp"
	"strings"
)

const response_header_annotation_key = "octoboy.ingress.kubernetes.io/add-response-header"

func init() {
	//tips nil 没办法强制转换成RewriteFilter，但是nil可以转换成*RewriteFilter，代表空指针
	registerFilterResponese(response_header_annotation_key, (*ResponseHeaderFilter)(nil))
}

type ResponseHeaderFilter struct {
	PathReg string
	Target  string
}

func (r *ResponseHeaderFilter) SetValue(value ...string) {
	r.Target = value[0]
}

func (r *ResponseHeaderFilter) SetPathReg(value ...string) {
}

func (r *ResponseHeaderFilter) Do(ctx *fasthttp.RequestCtx) {
	kvs := strings.Split(r.Target, ";")
	for _, kv := range kvs {
		set := strings.Split(kv, "=")
		ctx.Response.Header.Add(set[0], set[1])
	}

}
