package ingress

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
)

type IngressCtl struct {
	Svc *IngressSvc `inject:"-"`
}

func (this *IngressCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/ingress", this.ListAll)
	goft.Handle("POST", "/:cluster/ingress", this.PostIngress)
	goft.Handle("DELETE", "/:cluster/ingress", this.DeleteIngress)
}

func (this *IngressCtl) ListAll(c *gin.Context) goft.Json {
	ns := c.DefaultQuery("ns", "default")
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListAll(cluster, ns),
	}
}

func (this *IngressCtl) PostIngress(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	postModel := &IngressPost{}
	goft.Error(c.BindJSON(postModel))
	goft.Error(this.Svc.PostIngress(cluster, postModel))
	return gin.H{
		"code": 20000,
		"data": postModel,
	}
}

func (this *IngressCtl) DeleteIngress(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	name := c.Query("name")
	goft.Error(this.Svc.DeleteIngress(cluster, ns, name))
	return gin.H{
		"code":    20000,
		"message": "OK",
	}
}

func (this *IngressCtl) Name() string {
	return "IngressCtl"
}

func NewIngressCtl() *IngressCtl {
	return &IngressCtl{}
}
