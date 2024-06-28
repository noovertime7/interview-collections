package deployment

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type DeploymentCtl struct {
	Svc    *DeploySvc     `inject:"-"`
	K8sCli *models.CliMap `inject:"-"`
}

func (this *DeploymentCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/deployments", this.GetList)
	goft.Handle("GET", "/:cluster/deployments/:ns/:name", this.LoadDeploy)
	goft.Handle("POST", "/:cluster/deployments", this.SaveDeployment)
	goft.Handle("DELETE", "/:cluster/deployments/:ns/:name", this.RmDeployment)
}

func NewDeploymentCtl() *DeploymentCtl {
	return &DeploymentCtl{}
}

func (this *DeploymentCtl) Name() string {
	return "DeploymentCtl"
}

func (this *DeploymentCtl) GetList(ctx *gin.Context) goft.Json {
	ns := ctx.DefaultQuery("ns", "default")
	cluster := ctx.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListAll(cluster, ns),
	}
}

func (this *DeploymentCtl) LoadDeploy(ctx *gin.Context) goft.Json {
	ns := ctx.Param("ns")
	name := ctx.Param("name")
	cluster := ctx.Param("cluster")
	dep, err := this.Svc.GetDeployment(cluster, ns, name)
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": dep,
	}
}

func (this *DeploymentCtl) SaveDeployment(ctx *gin.Context) goft.Json {
	dep := &v1.Deployment{}
	goft.Error(ctx.ShouldBindJSON(dep))
	if ctx.Query("fast") != "" { //代表是快捷创建 。 要预定义一些值
		utils.InitLabel(dep)
	}
	update := ctx.Query("update") //代表是更新
	cluster := ctx.Param("cluster")
	if update != "" {
		_, err := (*this.K8sCli)[cluster].AppsV1().Deployments(dep.Namespace).Update(context.Background(), dep, v12.UpdateOptions{})
		goft.Error(err)
	} else {
		_, err := (*this.K8sCli)[cluster].AppsV1().Deployments(dep.Namespace).Create(context.Background(), dep, v12.CreateOptions{})
		goft.Error(err)
	}
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *DeploymentCtl) RmDeployment(ctx *gin.Context) goft.Json {
	ns := ctx.Param("ns")
	name := ctx.Param("name")
	cluster := ctx.Param("cluster")
	err := (*this.K8sCli)[cluster].AppsV1().Deployments(ns).Delete(context.Background(), name, v12.DeleteOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}
