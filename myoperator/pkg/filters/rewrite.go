package filters

import (
	"github.com/valyala/fasthttp"
	"regexp"
)

const rewrite_annotation_key = "octoboy.ingress.kubernetes.io/rewrite-target"

func init() {
	//tips nil 没办法强制转换成RewriteFilter，但是nil可以转换成*RewriteFilter，代表空指针
	registerFilter(rewrite_annotation_key, (*RewriteFilter)(nil))
}

type RewriteFilter struct {
	PathReg string
	Target  string
}

func (r *RewriteFilter) SetValue(value ...string) {
	r.Target = value[0]
}

func (r *RewriteFilter) SetPathReg(value ...string) {
	r.PathReg = value[0]
}

func (r *RewriteFilter) Do(ctx *fasthttp.RequestCtx) {
	uri := string(ctx.RequestURI())

	reg := regexp.MustCompile(r.PathReg)
	uri = reg.ReplaceAllString(uri, r.Target)

	ctx.Request.SetRequestURI(uri)

}

var _ ProxyFilter = &RewriteFilter{}
