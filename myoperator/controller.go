package main

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"jtproxy/pkg/filters"
	"jtproxy/pkg/kube"
	"jtproxy/pkg/sysinit"
	v1 "k8s.io/api/networking/v1"
	"log"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func ProxyHandler(ctx *fasthttp.RequestCtx) {
	//代表匹配到了 path
	if getProxy := sysinit.GetRoute(&ctx.Request); getProxy != nil {
		//要在这里 把请求重定向一下
		filters.ProxyFilters(getProxy.Filters).Do(ctx)
		getProxy.Proxy.ServeHTTP(ctx)
		filters.ProxyFilters(getProxy.FiltersResponese).Do(ctx) //修改响应头
	} else {
		ctx.Response.SetStatusCode(404)
		ctx.Response.SetBodyString("404...")
	}

}

// https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/builder#example-Builder
func main() {
	//首先从文件读取配置到内存
	err := sysinit.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	logf.SetLogger(zap.New())

	var log = logf.Log.WithName("jt-proxy")

	mgr, err := manager.New(kube.GetConfig(), manager.Options{})
	if err != nil {
		log.Error(err, "could not create manager")
		os.Exit(1)
	}

	//err = kube.SchemeBuilder.AddToScheme(mgr.GetScheme())
	//if err != nil {
	//	log.Error(err, "unable add schema")
	//	os.Exit(1)
	//} // ++ 添加mgr的scheme到schemebuilder中 用于crd
	c := kube.NewJtProxyController()

	err = builder.
		ControllerManagedBy(mgr). // 指定了manager
		For(&v1.Ingress{}).       // 指定要监听的cr
		Watches(
			&source.Kind{Type: &v1.Ingress{}},
			handler.Funcs{DeleteFunc: c.OnDelete},
		). // watch资源对象 触发 eventHandler
		//For(&kube.Route{}).                 // ++ 监听crd
		Complete(c) //指定了控制器
	if err != nil {
		log.Error(err, "could not create controller")
		os.Exit(1)
	}

	errChan := make(chan error)
	go func() {
		if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
			errChan <- err
		}
	}() // 启动controller，warning 启动后存量对象会触发reconcile，不停的去写文件，不适合生产环境
	//sysinit.InitConfig()
	go func() {
		if err := fasthttp.ListenAndServe(fmt.Sprintf(":%d", sysinit.SysConfig.Server.Port), ProxyHandler); err != nil {
			errChan <- err
		}
	}() //内置了一个反代软件，controller监听到ingress变化会reload反代的路由匹配
	if err = <-errChan; err != nil {
		log.Error(err, "could not start mgr or proxy.")
	}
}
