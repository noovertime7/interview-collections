package caches

import (
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
	"k8sapi-pro/src/models"
)

type ClusterHandler struct {
	h       *ResourceHandler
	Cluster string
}

func WrapperHandler(h *ResourceHandler, cluster string) *ClusterHandler {
	return &ClusterHandler{h: h, Cluster: cluster}
}

func (c ClusterHandler) OnAdd(obj interface{}) {
	c.h.OnAdd(obj, c.Cluster)
}

func (c ClusterHandler) OnUpdate(oldObj, newObj interface{}) {
	c.h.OnUpdate(oldObj, newObj, c.Cluster)
}

func (c ClusterHandler) OnDelete(obj interface{}) {
	c.h.OnDelete(obj)
}

// 入库资源对象通用handler
type ResourceHandler struct {
	Db *gorm.DB         `inject:"-"`
	Rm *meta.RESTMapper `inject:"-"`
}

func (r ResourceHandler) OnAdd(obj interface{}, cluster string) {
	if o, ok := obj.(runtime.Object); ok {
		if resource, err := models.NewResource(cluster, o, *r.Rm); err == nil {
			goft.Error(resource.Add(r.Db))
			//矛盾点:
			//1.有些前段数据是后端定制生成的，比如就绪副本数，不能通过取数据库原始数据来返回。
			//2.原来缓存方案是直接将缓存里的整个对象切片返回，现在还要请求一次数据库
			// 这两部分需要适配
			//wscore.ClientMap.SendAll(
			//	gin.H{
			//		"type": resource.Resource,
			//		"result": gin.H{
			//			"ns": resource.NameSpace,
			//			"data":
			//		}
			//})
		} else {
			goft.Error(err)
		}
	}
}

func (r ResourceHandler) OnUpdate(oldObj, newObj interface{}, cluster string) {
	if o, ok := newObj.(runtime.Object); ok {
		if resource, err := models.NewResource(cluster, o, *r.Rm); err == nil {
			goft.Error(resource.Update(r.Db))
		} else {
			goft.Error(err)
		}
	}
}

func (r ResourceHandler) OnDelete(obj interface{}) {
	if o, ok := obj.(runtime.Object); ok {
		metaobj, err := meta.Accessor(o) //tips 直接获取对象元数据信息，从其中取uid，不再使用序列化的反射消耗性能
		goft.Error(err)
		uid := metaobj.GetUID()
		goft.Error(models.Delete(string(uid), r.Db))
	}
}

func NewResourceHandler() *ResourceHandler {
	return &ResourceHandler{}
}

// var _ cache.ResourceEventHandler = ResourceHandler{}
var _ cache.ResourceEventHandler = ClusterHandler{}
