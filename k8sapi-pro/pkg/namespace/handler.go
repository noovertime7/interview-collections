package namespace

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/cache"
	"k8sapi-pro/src/utils"
)

type ClusteredNsHandler struct {
	h       *NsHandler
	cluster string
}

func WrapperNsHandler(h *NsHandler, cluster string) *ClusteredNsHandler {
	return &ClusteredNsHandler{h: h, cluster: cluster}
}

func (c *ClusteredNsHandler) OnAdd(obj interface{}) {
	c.h.OnAdd(obj, c.cluster)
}

func (c *ClusteredNsHandler) OnUpdate(oldObj, newObj interface{}) {
	c.h.OnUpdate(oldObj, newObj, c.cluster)
}

func (c *ClusteredNsHandler) OnDelete(obj interface{}) {
	c.h.OnDelete(obj, c.cluster)
}

type NsHandler struct {
	NsMap *NsMap `inject:"-"`
}

func (n *NsHandler) OnAdd(obj interface{}, cluster string) {
	if o, ok := obj.(runtime.Object); ok {
		if b, err := json.Marshal(o); err == nil {
			ns := utils.Convert(b)
			n.NsMap.Add(cluster, ns.(*v1.Namespace))
		}
	}
}

func (n *NsHandler) OnUpdate(oldObj, newObj interface{}, cluster string) {
	if o, ok := newObj.(runtime.Object); ok {
		if b, err := json.Marshal(o); err == nil {
			ns := utils.Convert(b)
			n.NsMap.Update(cluster, ns.(*v1.Namespace))
		}
	}
}

func (n *NsHandler) OnDelete(obj interface{}, cluster string) {
	if o, ok := obj.(runtime.Object); ok {
		if b, err := json.Marshal(o); err == nil {
			ns := utils.Convert(b)
			n.NsMap.Delete(cluster, ns.(*v1.Namespace))
		}
	}
}

// var _ cache.ResourceEventHandler = NsHandler{}
var _ cache.ResourceEventHandler = &ClusteredNsHandler{}
