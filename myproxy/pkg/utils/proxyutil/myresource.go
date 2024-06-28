package proxyutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

// 放内部资源
// 目前 放到  /apps/v1里
const (
	// 响应做修改时的正则
	myres_pattern = `/apis/apps/v1` //- // cluster 强制 归类于 apps

	MyResourceApiVersion = "apps/v1"
	MyClusterName        = "clusters"
	MyClusterKind        = "Cluster"
	MyClusterListKind    = "ClusterList"
	MyClusterShortName   = "cl"
)

// 把自定义资源 做了一些封装。 因为后面要支持多个
type MyResource struct {
	req    *http.Request
	writer http.ResponseWriter
}

func NewMyResource(req *http.Request, writer http.ResponseWriter) *MyResource {
	return &MyResource{req: req, writer: writer}
}

const cluster_pattern_handler = `/apis/apps/v1/namespaces/.*?/clusters`

// 这个函数是给 MyResource.HandlerForCluster 调用的
func (my *MyResource) handlerForCluster(clusters []string) []byte {
	r := regexp.MustCompile(cluster_pattern_handler)
	if r.MatchString(my.req.RequestURI) && my.req.Method == "GET" {
		// 构建一个 unstructured
		ret := &unstructured.UnstructuredList{
			Items: make([]unstructured.Unstructured, len(clusters)),
		}
		ret.SetKind(MyClusterListKind)
		ret.SetAPIVersion(MyResourceApiVersion)

		for i, cluster := range clusters {
			obj := unstructured.Unstructured{}
			obj.SetAPIVersion(MyResourceApiVersion)
			obj.SetKind(MyClusterKind)
			obj.SetName(cluster)
			obj.SetCreationTimestamp(metav1.NewTime(time.Now()))
			ret.Items[i] = obj
		}

		b, err := ret.MarshalJSON()
		if err != nil {
			log.Println(err)
			return nil
		}
		return b
	}
	return nil
}

// 这里引出一个思路 kubectl get xxx的操作其实是先获取api-resources的列表（已在transport中加工了结果）
// 如果客户端判断如果在返回的列表中，则根据返回的信息拼凑url再次请求apiserver
// 返回值 bool，代表是否拦截到，拦截到直接返回 。 如果是TRUE 外部则不应该继续响应
func (my *MyResource) HandlerForCluster(clusterMap map[string]http.Handler) bool {
	clusterNames := []string{}
	for k, _ := range clusterMap {
		clusterNames = append(clusterNames, k)
	}
	myres := my.handlerForCluster(clusterNames)
	if myres != nil {
		my.writer.Header().Set("Content-type", "application/json")
		my.writer.Header().Set("Content-Length", strconv.Itoa(len(myres)))
		my.writer.Write(myres)
		return true
	}
	return false
}

// 拦截自定义资源 做修改  ---在transport.go 里拦截
func handlerMyResource(obj *unstructured.Unstructured, req *http.Request) {
	r := regexp.MustCompile(myres_pattern)
	if r.MatchString(req.RequestURI) && obj.GetKind() == "APIResourceList" {
		if resList, ok := obj.Object["resources"].([]interface{}); ok {
			if !existsClusterDef(resList) {
				resList = append(resList, getClusterMap())
				obj.Object["resources"] = resList
			}

		}
	}
}

// 是否已经定义过 自定义资源
func existsClusterDef(resList []interface{}) bool {
	for _, res := range resList {
		if r, ok := res.(map[string]interface{}); ok {
			if n, ok := r["name"]; ok && n == MyClusterName {
				return true
			}
		}
	}
	return false
}

func getClusterMap() map[string]interface{} {
	return map[string]interface{}{
		"kind":               MyClusterKind,
		"name":               MyClusterName,
		"namespaced":         true,
		"singularName":       "",
		"storageVersionHash": "",
		"categories":         []string{"all"},
		"shortNames":         []string{MyClusterShortName},
		"verbs":              []string{"get", "list", "patch"},
	}
}
