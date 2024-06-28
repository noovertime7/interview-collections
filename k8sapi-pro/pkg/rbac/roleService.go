package rbac

import (
	"gorm.io/gorm"
	v1 "k8s.io/api/rbac/v1"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type RoleSvc struct {
	Db *gorm.DB `inject:"-"`
}

func (this *RoleSvc) ListRoles(cluster, ns string) (ret []*RoleModel) {
	if list, err := models.List(cluster, ns, "Role", this.Db); err == nil {
		ret = make([]*RoleModel, len(list))
		for i, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.Role)
			ret[i] = &RoleModel{
				Name:       obj.Name,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
				NameSpace:  obj.Namespace,
			}
		}
	}
	return
}

func (this *RoleSvc) ListClusterRoles(cluster string) (ret []*v1.ClusterRole) {
	if list, err := models.ListNoNamespaced(cluster, "ClusterRole", this.Db); err == nil {
		ret = make([]*v1.ClusterRole, len(list))
		for i, resource := range list {
			ret[i] = utils.Convert([]byte(resource.Object)).(*v1.ClusterRole)
		}
	}
	return
}

func (this *RoleSvc) GetRole(cluster string, ns string, name string) *v1.Role {
	if r, err := models.Take(cluster, ns, name, "Role", this.Db); err == nil {
		return utils.Convert([]byte(r.Object)).(*v1.Role)
	}
	return nil
}

func (this *RoleSvc) GetClusterRole(cluster string, name string) *v1.ClusterRole {
	if r, err := models.TakeNoNamespaced(cluster, name, "ClusterRole", this.Db); err == nil {
		return utils.Convert([]byte(r.Object)).(*v1.ClusterRole)
	}
	return nil
}

func (this *RoleSvc) ListRoleBindings(cluster string, ns string) (ret []*RoleBindingModel) {
	if list, err := models.List(cluster, ns, "RoleBinding", this.Db); err == nil {
		ret = make([]*RoleBindingModel, len(list))
		for i, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.RoleBinding)
			ret[i] = &RoleBindingModel{
				Name:       obj.Name,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
				NameSpace:  obj.Namespace,
				Subject:    obj.Subjects,
				RoleRef:    obj.RoleRef,
			}
		}
	}
	return
}

func (this *RoleSvc) GetRoleBinding(cluster, ns, name string) *v1.RoleBinding {
	if r, err := models.Take(cluster, ns, name, "RoleBinding", this.Db); err == nil {
		return utils.Convert([]byte(r.Object)).(*v1.RoleBinding)
	}
	return nil
}

func (this *RoleSvc) ListClusterRoleBindings(cluster string) (ret []*v1.ClusterRoleBinding) {
	if list, err := models.ListNoNamespaced(cluster, "ClusterRoleBinding", this.Db); err == nil {
		ret = make([]*v1.ClusterRoleBinding, len(list))
		for i, resource := range list {
			ret[i] = utils.Convert([]byte(resource.Object)).(*v1.ClusterRoleBinding)
		}
	}
	return
}

func (this *RoleSvc) GetClusterRoleBinding(cluster, name string) *v1.ClusterRoleBinding {
	if r, err := models.TakeNoNamespaced(cluster, name, "ClusterRoleBinding", this.Db); err == nil {
		return utils.Convert([]byte(r.Object)).(*v1.ClusterRoleBinding)
	}
	return nil
}
