package namespace

import (
	"gorm.io/gorm"
	"k8sapi-pro/src/models"
)

type NamespaceSvc struct {
	Db *gorm.DB `inject:"-"`
}

func (this *NamespaceSvc) List(cluster string) (ret []*Ns) {
	if resources, err := models.ListNoNamespaced(cluster, "Namespace", this.Db); err == nil {
		ret = make([]*Ns, len(resources))
		for i, resource := range resources {
			ret[i] = &Ns{Name: resource.Name}
		}
	}
	return
}
