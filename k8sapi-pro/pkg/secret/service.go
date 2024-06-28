package secret

import (
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type SecretSvc struct {
	Db *gorm.DB `inject:"-"`
}

// 前台用于显示Secret列表
func (this *SecretSvc) ListSecret(cluster, ns string) (ret []*SecretModel) {
	if list, err := models.List(cluster, ns, "Secret", this.Db); err == nil {
		for _, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.Secret)
			ret = append(ret, &SecretModel{
				Name:       obj.Name,
				NameSpace:  obj.Namespace,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
				Type:       SECRET_TYPE[string(obj.Type)], // 类型的翻译
			})
		}
	}
	return
}

// 解析 （如类型是 tls 的secret)
func (this *SecretSvc) ParseIfTLS(t string, data map[string][]byte) interface{} {
	if t == "kubernetes.io/tls" {
		if crt, ok := data["tls.crt"]; ok {
			crtModel := utils.ParseCert(crt)
			if crtModel != nil {
				return crtModel
			}
		}
	}
	return struct{}{}

}
