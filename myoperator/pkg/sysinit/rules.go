package sysinit

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/valyala/fasthttp"
	"github.com/yeqown/fasthttp-reverse-proxy/v2"
	"jtproxy/pkg/filters"
	v1 "k8s.io/api/networking/v1"
	"net/http"
	"net/url"
)

type ProxyHandler struct {
	Proxy            *proxy.ReverseProxy // proxy对象。 保存proxy
	FiltersResponese []filters.ProxyFilter
	Filters          []filters.ProxyFilter
}

// 空函数没啥用
func (this *ProxyHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

// 解析配置文件中的rules， 初始化 路由
func ParseRule() {
	//现在要循环 遍历
	for _, ingress := range SysConfig.Ingress {
		for _, rule := range ingress.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				//构建 反代对象
				rProxy := proxy.NewReverseProxy(
					fmt.Sprintf("%s:%d", path.Backend.Service.Name, path.Backend.Service.Port.Number))

				routeBud := NewRouteBuilder()

				routeBud.
					SetPath(path.Path, path.PathType != nil && *path.PathType == v1.PathTypeExact).
					SetHost(rule.Host, rule.Host != "").
					Build(&ProxyHandler{
						Proxy:            rProxy,
						Filters:          filters.CheckAnnotations(ingress.Annotations, true),
						FiltersResponese: filters.CheckAnnotations(ingress.Annotations, false),
					})

			}
		}
	}

}

// 获取路由   （先匹配 请求path ，如果匹配到 ，会返回 对应的proxy 对象)
func GetRoute(req *fasthttp.Request) *ProxyHandler {
	match := &mux.RouteMatch{}
	httpReq := &http.Request{
		URL:    &url.URL{Path: string(req.URI().Path())},
		Method: string(req.Header.Method()),
		Host:   string(req.Header.Host()),
	}
	if MyRouter.Match(httpReq, match) {
		proxyHandler := match.Handler.(*ProxyHandler)
		//
		pathReg, err := match.Route.GetPathRegexp()
		if err == nil {
			filters.ProxyFilters(proxyHandler.Filters).SetPathReg(pathReg)
			filters.ProxyFilters(proxyHandler.FiltersResponese).SetPathReg(pathReg)
		} //从match获取到path的正则表达式 赛到filter里去，用于reg包的正则匹配
		return proxyHandler
	}
	return nil
}
