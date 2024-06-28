package proxyutil

import (
	"fmt"
	"net/http"
	"regexp"
)

// keypoint 嵌套 这里的思想很重要，类似于中间件，在执行serve前拦截
// http包中有一个handler接口实现了serverHTTP方法，包装一层即可实现请求前拦截
type MyProxyHandler struct {
	//http.Handler
	handlers   map[string]http.Handler
	defaultCtx string
}

const OpenApiPattern = "/openapi/v[2|3]"

func (h *MyProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	//获取req header 鉴权,并记录日志
	if verifyPass := Auth(req); !verifyPass {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	//  kubectl get cluster，拦截下来，自己包装结果返回 （需要先在获取api-resource的结果中加入假的资源，因为没有对应cr）
	if NewMyResource(req, w).HandlerForCluster(h.handlers) {
		return
	}

	//openapi请求直接转发
	if regexp.MustCompile(OpenApiPattern).MatchString(req.RequestURI) {
		fmt.Println("openapi")
		h.handlers[h.defaultCtx].ServeHTTP(w, req)
		return
	}

	//获取集群名称
	cluster := parseCluster(req)
	if cluster == "" {
		cluster = h.defaultCtx
	}

	req.Header.Set("from_cluster", cluster) //塞入集群名称，runTrip中获得并加入到返回结果中
	req.Header.Del("Authorization")         //访问aws集群需要去除

	//fmt.Println("cluster是" + cluster)
	//fmt.Println("url是" + req.URL.String())
	if h.handlers[cluster] != nil {
		h.handlers[cluster].ServeHTTP(w, req)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

	//用于请求多集群并合并返回结果
	//mywriter := WrapWriter()
	//req.Header.Add("from_cluster", "天翼云")
	//req2 := cloneRequest(req, "eks") // 必须要在 serveHttp之前 克隆
	//h.handlers["e-context"].ServeHTTP(mywriter, req)
	//
	//h.handlers["eks-test"].ServeHTTP(mywriter, req2)
	//
	////合并响应写入真实writer
	//writeResponse(w, mywriter)
}

// 单集群handler
type SingleHandler struct {
	http.Handler
}

func (h *SingleHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//中间件的实现,这里是执行前拦截。拦截响应的话，需要对transport做包装处理
	//todo 指标 以url为key 供prom采集
	//fmt.Println(req.URL)
	//fmt.Println(req.Header) //todo 获取token做权限处理
	h.ServeHTTP(w, req)
}

func WrapperProxyHandler(h http.Handler) *SingleHandler {
	return &SingleHandler{
		Handler: h,
	}
}
