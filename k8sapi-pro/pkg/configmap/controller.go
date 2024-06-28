package configmap

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sapi-pro/src/models"
)

type ConfigMapCtl struct {
	Svc    *ConfigMapSvc  `inject:"-"`
	CliMap *models.CliMap `inject:"-"`
}

func (this *ConfigMapCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/configmaps", this.ListAll)
	goft.Handle("DELETE", "/:cluster/configmaps", this.RmConfigMap)
	goft.Handle("POST", "/:cluster/configmaps", this.PostConfigMap)
	goft.Handle("GET", "/:cluster/configmaps/:ns/:name", this.Detail)
}

func (this *ConfigMapCtl) ListAll(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListConfigMap(cluster, ns), //暂时 不分页
	}
}

func (this *ConfigMapCtl) RmConfigMap(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	name := c.DefaultQuery("name", "")
	goft.Error((*this.CliMap)[cluster].CoreV1().ConfigMaps(ns).Delete(context.Background(), name, metav1.DeleteOptions{}))
	return gin.H{
		"code":    20000,
		"message": "OK",
	}
}

func (this *ConfigMapCtl) PostConfigMap(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	model := &PostConfigMap{}
	goft.Error(c.BindJSON(model))
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Name,
			Namespace: model.NameSpace,
		},
		Data: model.Data,
	}
	if model.IsUpdate {
		_, err := (*this.CliMap)[cluster].CoreV1().ConfigMaps(model.NameSpace).Update(context.Background(), cm, metav1.UpdateOptions{})
		goft.Error(err)
	} else {
		_, err := (*this.CliMap)[cluster].CoreV1().ConfigMaps(model.NameSpace).Create(context.Background(), cm, metav1.CreateOptions{})
		goft.Error(err)
	}
	return gin.H{
		"code":    20000,
		"message": "OK",
	}
}

func (this *ConfigMapCtl) Detail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.Param("ns")
	name := c.Param("name")
	cm, err := (*this.CliMap)[cluster].CoreV1().ConfigMaps(ns).Get(context.Background(), name, metav1.GetOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": ConfigMapModel{
			Name:       cm.Name,
			NameSpace:  cm.Namespace,
			CreateTime: cm.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Content:    cm.Data,
		},
	}
}

func (this *ConfigMapCtl) Name() string {
	return "ConfigMapCtl"
}

func NewConfigMapCtl() *ConfigMapCtl {
	return &ConfigMapCtl{}
}
