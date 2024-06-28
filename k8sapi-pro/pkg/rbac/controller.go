package rbac

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sapi-pro/src/models"
)

type RBACCtl struct {
	RoleService *RoleSvc       `inject:"-"`
	SaService   *SaSvc         `inject:"-"`
	CliMap      *models.CliMap `inject:"-"`
}

func (this *RBACCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/clusterroles", this.ClusterRoles)
	goft.Handle("DELETE", "/:cluster/clusterroles", this.DeleteClusterRole)
	goft.Handle("POST", "/:cluster/clusterroles", this.CreateClusterRole) //创建集群角色
	goft.Handle("GET", "/:cluster/clusterroles/:cname", this.ClusterRolesDetail)
	goft.Handle("POST", "/:cluster/clusterroles/:cname", this.UpdateClusterRolesDetail)

	goft.Handle("GET", "/:cluster/clusterrolebindings", this.ClusterRoleBindingList)
	goft.Handle("POST", "/:cluster/clusterrolebindings", this.CreateClusterRoleBinding)
	goft.Handle("PUT", "/:cluster/clusterrolebindings", this.AddUserToClusterRoleBinding)
	goft.Handle("DELETE", "/:cluster/clusterrolebindings", this.DeleteClusterRoleBinding)

	goft.Handle("GET", "/:cluster/roles", this.Roles)
	goft.Handle("GET", "/:cluster/role/:ns/:name", this.RoleDetail)
	goft.Handle("POST", "/:cluster/roles/:ns/:rolename", this.UpdateRolesDetail) //修改角色
	goft.Handle("GET", "/:cluster/rolebindings", this.RoleBindingList)
	goft.Handle("POST", "/:cluster/rolebindings", this.CreateRoleBinding)
	goft.Handle("DELETE", "/:cluster/rolebindings", this.DeleteRoleBinding)
	goft.Handle("POST", "/:cluster/roles", this.CreateRole)
	goft.Handle("DELETE", "/:cluster/roles", this.DeleteRole)
	goft.Handle("PUT", "/:cluster/rolebindings", this.AddUserToRoleBinding) //添加用户到binding

	goft.Handle("GET", "/:cluster/sa", this.SaList)
}

func (this *RBACCtl) Roles(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.RoleService.ListRoles(cluster, ns),
	}
}

func (this *RBACCtl) ClusterRoles(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.RoleService.ListClusterRoles(cluster),
	}
}

func (this *RBACCtl) DeleteClusterRole(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	name := c.DefaultQuery("name", "")
	err := (*this.CliMap)[cluster].RbacV1().ClusterRoles().Delete(context.Background(), name, metav1.DeleteOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) RoleDetail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.Param("ns")
	name := c.Param("name")
	return gin.H{
		"code": 20000,
		"data": this.RoleService.GetRole(cluster, ns, name),
	}
}

// 更新角色
func (this *RBACCtl) UpdateRolesDetail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.Param("ns")
	rname := c.Param("rolename")
	role := this.RoleService.GetRole(cluster, ns, rname)
	postRole := rbacv1.Role{}
	goft.Error(c.ShouldBindJSON(&postRole)) //获取提交过来的对象

	role.Rules = postRole.Rules //目前修改只允许修改 rules，其他不允许。大家可以自行扩展，如标签也允许修改
	_, err := (*this.CliMap)[cluster].RbacV1().Roles(role.Namespace).Update(context.Background(), role, metav1.UpdateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) RoleBindingList(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.RoleService.ListRoleBindings(cluster, ns),
	}
}

func (this *RBACCtl) DeleteRole(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	name := c.DefaultQuery("name", "")
	err := (*this.CliMap)[cluster].RbacV1().Roles(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}
func (this *RBACCtl) CreateRole(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	role := rbacv1.Role{} //原生的k8s role 对象
	goft.Error(c.ShouldBindJSON(&role))
	role.APIVersion = "rbac.authorization.k8s.io/v1"
	role.Kind = "Role"
	_, err := (*this.CliMap)[cluster].RbacV1().Roles(role.Namespace).Create(context.Background(), &role, metav1.CreateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) CreateClusterRole(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	clusterRole := rbacv1.ClusterRole{}
	goft.Error(c.ShouldBindJSON(&clusterRole))
	// 不管是自定义的model还是前端根据结构体传入的 kind 和 APIversion 这部分都是缺失的  需要手动填充一下
	clusterRole.APIVersion = "rbac.authorization.k8s.io/v1"
	clusterRole.Kind = "ClusterRole"
	_, err := (*this.CliMap)[cluster].RbacV1().ClusterRoles().Create(context.Background(), &clusterRole, metav1.CreateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

// //获取集群角色详细
func (this *RBACCtl) ClusterRolesDetail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	rname := c.Param("cname") //集群角色名
	return gin.H{
		"code": 20000,
		"data": this.RoleService.GetClusterRole(cluster, rname),
	}
}

// 更新集群角色
func (this *RBACCtl) UpdateClusterRolesDetail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	cname := c.Param("cname") //集群角色名
	clusterRole := this.RoleService.GetClusterRole(cluster, cname)
	postRole := rbacv1.ClusterRole{}
	goft.Error(c.ShouldBindJSON(&postRole)) //获取提交过来的对象

	clusterRole.Rules = postRole.Rules //目前修改只允许修改 rules，其他不允许。大家可以自行扩展，如标签也允许修改
	_, err := (*this.CliMap)[cluster].RbacV1().ClusterRoles().Update(context.Background(), clusterRole, metav1.UpdateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) DeleteRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	name := c.DefaultQuery("name", "")
	err := (*this.CliMap)[cluster].RbacV1().RoleBindings(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) CreateRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	rb := &rbacv1.RoleBinding{}
	goft.Error(c.ShouldBindJSON(rb))
	_, err := (*this.CliMap)[cluster].RbacV1().RoleBindings(rb.Namespace).Create(context.Background(), rb, metav1.CreateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) AddUserToRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.Query("ns")
	name := c.Query("name")
	t := c.DefaultQuery("type", "")
	subject := rbacv1.Subject{} // 传过来
	goft.Error(c.ShouldBindJSON(&subject))
	if subject.Kind == "ServiceAccount" {
		subject.APIGroup = ""
	}
	rb := this.RoleService.GetRoleBinding(cluster, ns, name) //通过名称获取 rolebinding对象
	if t != "" {                                             //代表删除

		for i, sub := range rb.Subjects {
			if sub.Kind == subject.Kind && sub.Name == subject.Name {
				rb.Subjects = append(rb.Subjects[:i], rb.Subjects[i+1:]...)
				break //确保只删一个（哪怕有同名同kind用户)
			}
		}
		fmt.Println(rb.Subjects)
	} else {
		rb.Subjects = append(rb.Subjects, subject)
	}
	_, err := (*this.CliMap)[cluster].RbacV1().RoleBindings(ns).Update(context.Background(), rb, metav1.UpdateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "OK",
	}
}

func (this *RBACCtl) SaList(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	return gin.H{
		"code": 20000,
		"data": this.SaService.ListSa(cluster, ns),
	}
}

func (this *RBACCtl) ClusterRoleBindingList(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.RoleService.ListClusterRoleBindings(cluster),
	}
}

func (this *RBACCtl) AddUserToClusterRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	name := c.DefaultQuery("name", "") //clusterrolebinding 名称
	t := c.DefaultQuery("type", "")    //如果没传值就是增加，传值（不管什么代表删除)
	subject := rbacv1.Subject{}        // 传过来
	goft.Error(c.ShouldBindJSON(&subject))
	if subject.Kind == "ServiceAccount" {
		subject.APIGroup = ""
	}
	rb := this.RoleService.GetClusterRoleBinding(cluster, name) //通过名称获取 clusterrolebinding对象
	if t != "" {                                                //代表删除
		for i, sub := range rb.Subjects {
			if sub.Kind == subject.Kind && sub.Name == subject.Name {
				rb.Subjects = append(rb.Subjects[:i], rb.Subjects[i+1:]...)
				break //确保只删一个（哪怕有同名同kind用户)
			}
		}
	} else {
		rb.Subjects = append(rb.Subjects, subject)
	}
	_, err := (*this.CliMap)[cluster].RbacV1().ClusterRoleBindings().Update(context.Background(), rb, metav1.UpdateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) DeleteClusterRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	name := c.DefaultQuery("name", "")
	err := (*this.CliMap)[cluster].RbacV1().ClusterRoleBindings().Delete(context.Background(), name, metav1.DeleteOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) CreateClusterRoleBinding(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	rb := &rbacv1.ClusterRoleBinding{}
	goft.Error(c.ShouldBindJSON(rb))
	_, err := (*this.CliMap)[cluster].RbacV1().ClusterRoleBindings().Create(context.Background(), rb, metav1.CreateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *RBACCtl) Name() string {
	return "RBACCtl"
}

func NewRBACCtl() *RBACCtl {
	return &RBACCtl{}
}
