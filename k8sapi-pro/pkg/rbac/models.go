package rbac

import "k8s.io/api/rbac/v1"

type RoleModel struct {
	Name       string
	NameSpace  string
	CreateTime string
}
type RoleBindingModel struct {
	Name       string
	NameSpace  string
	CreateTime string
	Subject    []v1.Subject //包含了 绑定用户 数据
	RoleRef    v1.RoleRef
}
