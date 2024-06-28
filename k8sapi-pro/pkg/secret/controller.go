package secret

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sapi-pro/src/models"
)

type SecretCtl struct {
	Svc    *SecretSvc     `inject:"-"`
	CliMap *models.CliMap `inject:"-"`
}

func (this *SecretCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/secrets", this.ListAll)
	goft.Handle("DELETE", "/:cluster/secrets", this.RmSecret)
	goft.Handle("POST", "/:cluster/secrets", this.PostSecret)
	goft.Handle("GET", "/:cluster/secrets/:ns/:name", this.Detail)
}

func (this *SecretCtl) RmSecret(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	name := c.DefaultQuery("name", "")
	goft.Error((*this.CliMap)[cluster].CoreV1().Secrets(ns).
		Delete(context.Background(), name, metav1.DeleteOptions{}))
	return gin.H{
		"code": 20000,
		"data": "OK",
	}
}
func (this *SecretCtl) ListAll(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListSecret(cluster, ns), //暂时 不分页
	}
}

func (this *SecretCtl) PostSecret(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	model := &PostSecret{}
	goft.Error(c.BindJSON(model))
	_, err := (*this.CliMap)[cluster].CoreV1().Secrets(model.NameSpace).Create(context.Background(),
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      model.Name,
				Namespace: model.NameSpace,
			},
			StringData: model.Data,
			Type:       corev1.SecretType(model.Type),
		}, metav1.CreateOptions{})
	goft.Error(err)
	return gin.H{
		"code":    20000,
		"message": "ok",
	}
}

func (this *SecretCtl) Detail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.Param("ns")
	name := c.Param("name")
	secret, err := (*this.CliMap)[cluster].CoreV1().Secrets(ns).Get(context.Background(), name, metav1.GetOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": SecretModel{
			Name:       secret.Name,
			NameSpace:  secret.Namespace,
			CreateTime: secret.CreationTimestamp.Format("2006-01-02 15:04:05"),
			Type:       string(secret.Type),
			Content:    secret.Data,
			ExtData:    this.Svc.ParseIfTLS(string(secret.Type), secret.Data),
		},
	}
}

func (this *SecretCtl) Name() string {
	return "SecretCtl"
}

func NewSecretCtl() *SecretCtl {
	return &SecretCtl{}
}
