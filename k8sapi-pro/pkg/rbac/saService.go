package rbac

import (
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type SaSvc struct {
	Db *gorm.DB `inject:"-"`
}

func (this *SaSvc) ListSa(cluster, ns string) (ret []*v1.ServiceAccount) {
	if list, err := models.List(cluster, ns, "ServiceAccount", this.Db); err == nil {
		ret = make([]*v1.ServiceAccount, len(list))
		for i, resource := range list {
			ret[i] = utils.Convert([]byte(resource.Object)).(*v1.ServiceAccount)
		}
	}
	return
}
