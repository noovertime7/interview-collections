package kube

import (
	"context"
	"jtproxy/pkg/sysinit"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/util/workqueue"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type JtProxyController struct {
	client.Client
}

func NewJtProxyController() *JtProxyController {
	return &JtProxyController{}
}

func (a *JtProxyController) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	resource := &v1.Ingress{}
	//resource := &Route{} // 监听crd
	err := a.Client.Get(ctx, req.NamespacedName, resource) //发布后的ingress资源，都会通过控制器接收到
	if err != nil {
		return reconcile.Result{}, err
	}
	if a.IsJtProxy(resource.Annotations) {
		err = sysinit.UpdateConfig(resource)
		if err != nil {
			return reconcile.Result{}, err
		}
	} //根据注解筛选

	return reconcile.Result{}, nil
}

//https://github.com/kubernetes-sigs/controller-runtime/tree/master/examples/crd
// 监听crd，可以参考官方案例的实现，需要使用1.(schemeBuilder，定义类型，实现deepcopy) 2.在mgr中指定监听资源和加入scheme

func (a *JtProxyController) InjectClient(c client.Client) error {
	a.Client = c
	return nil
} //非常简单的实现 启动时由manager创建client并赋值进来

func (a *JtProxyController) IsJtProxy(annotations map[string]string) bool {
	if annotations["kubernetes.io/ingress.class"] == "octoboy" {
		return true
	}
	return false
}

func (a *JtProxyController) OnDelete(event event.DeleteEvent, r workqueue.RateLimitingInterface) {
	o := event.Object
	if a.IsJtProxy(o.GetAnnotations()) {
		err := sysinit.DeleteIngress(o.GetName(), o.GetNamespace())
		if err != nil {
			log.Println(err)
		}
	}
}
