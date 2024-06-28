package resources

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"k8sapi-pro/src/models"
	"sort"
	"strings"
)

type ResourcesCtl struct {
	CliMap *models.CliMap `inject:"-"`
}

func (this *ResourcesCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/v1/:cluster/resources", this.ListResources)
	goft.Handle("GET", "/clusters", this.ListCluster)
}

func (this *ResourcesCtl) ListCluster(c *gin.Context) goft.Json {
	res := []string{}
	for cluster, _ := range *this.CliMap {
		res = append(res, cluster)
	}
	sort.Sort(clusterList(res))
	return gin.H{
		"code": 20000,
		"data": res,
	}
}

func (this *ResourcesCtl) GetGroupVersion(str string) (group, version string) {
	list := strings.Split(str, "/")
	if len(list) == 1 {
		return "core", list[0]
	} else if len(list) == 2 {
		return list[0], list[1]
	}
	panic(any("error GroupVersion" + str))
}

// 获取所有资源
func (this *ResourcesCtl) ListResources(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	_, resList, err := (*this.CliMap)[cluster].ServerGroupsAndResources()
	goft.Error(err)
	res := make([]*GroupResources, 0)
	for _, r := range resList {
		group, version := this.GetGroupVersion(r.GroupVersion)
		resource := make([]*Resources, 0)
		for _, rr := range r.APIResources {
			resource = append(resource, &Resources{
				Name:  rr.Name,
				Verbs: rr.Verbs,
			})
		}
		res = append(res, &GroupResources{
			Group:     group,
			Version:   version,
			Resources: resource,
		})
	}
	return gin.H{
		"code": 20000,
		"data": res,
	}
}
func (*ResourcesCtl) Name() string {
	return "Resources"
}

func NewResourcesCtl() *ResourcesCtl {
	return &ResourcesCtl{}
}
