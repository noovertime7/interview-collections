package service

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
)

type ServiceCtl struct {
	Svc *ServiceSvc `inject:"-"`
}

func (this *ServiceCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/svc", this.ListAll)
}

func (this *ServiceCtl) ListAll(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListAll(cluster, ns),
	}
}

func (this *ServiceCtl) Name() string {
	return "ServiceCtl"
}

func NewServiceCtl() *ServiceCtl {
	return &ServiceCtl{}
}
