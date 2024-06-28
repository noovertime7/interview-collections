package configmap

import (
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type ConfigMapSvc struct {
	Db *gorm.DB `inject:"-"`
}

// 前台用于显示列表
func (this *ConfigMapSvc) ListConfigMap(cluster, ns string) (ret []*ConfigMapModel) {
	if list, err := models.List(cluster, ns, "ConfigMap", this.Db); err == nil {
		for _, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.ConfigMap)
			ret = append(ret, &ConfigMapModel{
				Name:       obj.Name,
				NameSpace:  obj.Namespace,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return
}
