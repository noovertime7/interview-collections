package deployment

import (
	"fmt"
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/gorm"
	v1 "k8s.io/api/apps/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type DeploySvc struct {
	Db *gorm.DB `inject:"-"`
}

func NewDeploySvc() *DeploySvc {
	return &DeploySvc{}
}

func (this *DeploySvc) ListAll(cluster, ns string) (ret []*Deployment) {
	if list, err := models.List(cluster, ns, "Deployment", this.Db); err == nil { //db查询
		for _, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.Deployment) //转换对象
			ret = append(ret, &Deployment{Name: obj.Name,
				NameSpace:   obj.Namespace,
				Replicas:    [3]int32{obj.Status.Replicas, obj.Status.AvailableReplicas, obj.Status.UnavailableReplicas},
				Images:      utils.GetImages(*obj),
				IsCompleted: utils.IsCompleted(obj),
				Message:     utils.GetAvailableMessage(obj),
				CreateTime:  obj.CreationTimestamp.Format("2006-01-02 - 15:04:05"),
			})
		}
	} else {
		goft.Error(err)
	}
	return
}

func (this *DeploySvc) GetDeployment(cluster, ns, name string) (*v1.Deployment, error) {
	res := &v1.Deployment{}
	if r, err := models.Take(cluster, ns, name, "Deployment", this.Db); err == nil {
		if deploy, ok := utils.Convert([]byte(r.Object)).(*v1.Deployment); ok {
			res = deploy
		} else {
			return nil, fmt.Errorf("type error! name:%s, namespace: %s.", name, ns)
		}
	}
	return res, nil
}
