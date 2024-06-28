package service

import (
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type ServiceSvc struct {
	Db *gorm.DB `inject:"-"`
}

func (this *ServiceSvc) ListAll(cluster, ns string) (ret []*ServiceModel) {
	if list, err := models.List(cluster, ns, "Service", this.Db); err == nil {
		for _, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.Service)
			ret = append(ret, &ServiceModel{
				Name:       obj.Name,
				NameSpace:  obj.Namespace,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return
}
