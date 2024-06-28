package namespace

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
)

type NsCtl struct {
	//NsMap *NsMap        `inject:"-"`
	Svc *NamespaceSvc `inject:"-"`
}

func (this *NsCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/namespace", this.ListAll)
}

func (this *NsCtl) ListAll(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.Svc.List(cluster),
	}
}

func (n NsCtl) Name() string {
	return "NsCtl"
}

func NewNsCtl() *NsCtl {
	return &NsCtl{}
}
