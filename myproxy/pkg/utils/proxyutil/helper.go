package proxyutil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/apimachinery/pkg/util/yaml"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	ClusterParam = "labelSelector" //原生的selector用来获取有对应标签的资源
	ClusterKey   = "cluster"
)

// 解析 app=ngx,cluster=xxx 的字符串
// string 是cluster的值
func parseSelectorIfCluster(param string) string {
	pair := strings.Split(param, "=")
	if len(pair) == 2 {
		if pair[0] == ClusterKey {
			return pair[1]
		}
	}
	return ""
}

const ClusterAnnotation = "octoboy/cluster"

// keypoint kubectl apply时会先发起一个get请求，根据返回结果来post或者patch
func parseCluster(req *http.Request) string {

	if req.Method == "POST" { //POST对象的url里面没有对象的名称 需要从body中的name获取，然后将集群名称返回
		// resourceName mycm.cluster.eks-test -> mycm
		cluster := replaceForKubectlPost(req)
		return cluster
	}

	//GET（单个对象），PATCH需要从param里面取对象的名称并做切割
	replacedPath, cluster, isreplace := stripClusterStr(req.URL.Path)
	if isreplace {
		req.URL.RawPath = replacedPath
		req.URL.Path = replacedPath

		// patch同样需要修改body里面的Name，但和post的body不同
		if req.Method == "PATCH" {
			replaceForKubectlPatch(req)
		}

		return cluster
	}

	//GET列表对象
	if req.Method == "GET" {
		cluster = parseClusterFromQuery(req)
	}
	return cluster
}

// GET请求(请求列表)的处理 ，请求单个对象的通过资源对象名称解析，因为kubectl apply会自动发起一个无法夹带参数的get请求
// aaa?a=1&b=3&c=4
func parseClusterFromQuery(req *http.Request) string {
	cluster := ""
	newSelector := []string{}

	values := req.URL.Query() //保存个副本
	//解析完成后，要去掉cluster=xxx  ，重新设置 参数，否则会查不到
	if selector := req.URL.Query().Get(ClusterParam); selector != "" {
		//按逗号切割
		strSplit := strings.Split(selector, ",")
		for _, param := range strSplit {

			if c := parseSelectorIfCluster(param); c != "" {
				cluster = c
			} else {
				newSelector = append(newSelector, param) //把不是cluster的参数 依然加入
			}
		}
		//更新request的url --- 排除cluster标记

		values.Set(ClusterParam, strings.Join(newSelector, ","))
		req.URL.RawQuery = values.Encode()
	}
	return cluster
}

// 返回集群名称
func replaceForKubectlPost(req *http.Request) string {
	// 获取非结构化对象
	tmpReq := cloneRequest(req, "")
	b, err := ioutil.ReadAll(tmpReq.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	obj := &unstructured.Unstructured{}
	err = yaml.Unmarshal(b, obj)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	// 重新获取并设置对象名称
	replacedName, cluster, _ := stripClusterStr(obj.GetName())
	obj.SetName(replacedName)
	// 重新打入body
	b, _ = obj.MarshalJSON()
	req.Body = ioutil.NopCloser(bytes.NewReader(b))
	req.ContentLength = int64(len(b))
	return cluster
}

/*
*
patch的示例内容
{"annotations":null,"metadata":{"annotations":{"kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"
v1\",\"data\":{\"name\":\"shenyi3\"},\"kind\":\"ConfigMap\",\"metadata\":{\"annotations\":{},\"name\":\"mycm.cluster.
aliyun\",\"namespace\":\"default\"}}\n"},"name":"mycm.cluster.aliyun"},"name":null,"namespace":null}
*/
func replaceForKubectlPatch(req *http.Request) {
	tmpReq := cloneRequest(req, "")
	b, err := ioutil.ReadAll(tmpReq.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	vmap := make(map[string]interface{})

	if err := json.Unmarshal(b, &vmap); err == nil {
		if metadata, ok := vmap["metadata"]; ok {

			if metadata_map, ok := metadata.(map[string]interface{}); ok {

				//替换本身的name
				if old_name, ok := metadata_map["name"]; ok {
					replacedName, _, isreplace := stripClusterStr(old_name.(string))
					if isreplace {
						metadata_map["name"] = replacedName
					}
				}

				//重新覆盖
				vmap["metadata"] = metadata_map

				//修改http body
				b, _ = json.Marshal(vmap)
				req.Body = ioutil.NopCloser(bytes.NewReader(b))
				req.ContentLength = int64(len(b))
			}
		}
	}
}

// 优先级最高的 ,名称只支持
const ClusterUrlPattern = `\.cluster\.([a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*?)`

// 为了不影响原生的 name。 需要把 .cluster.xxx给替换掉
// 返回值有两个。 string(1)=替换过后的值，string(2)=匹配到的cluster值 bool=是否产生了替换
func stripClusterStr(str string) (string, string, bool) {
	reg := regexp.MustCompile(ClusterUrlPattern)
	if reg.MatchString(str) { //匹配到
		matchResult := reg.FindStringSubmatch(str)
		if len(matchResult) < 2 {
			return str, "", false
		}
		str = strings.Replace(str, ".cluster."+matchResult[1], "", -1)
		return str, matchResult[1], true
	}
	return str, "", false
}

// Deprecated: 解析annotation
func parseClusterFromAnnotation(req *http.Request) string {

	cluster := ""

	//由于 kubectl apply 有 --selector的args入参, 所以需要通过其他方式来解析集群标示，这里采用解析 annotation的方法
	if req.Method == "POST" || req.Method == "PATCH" { // apply 的时候是POST 或patch 。我们要判断注解

		tmpReq := cloneRequest(req, "")
		b, err := ioutil.ReadAll(tmpReq.Body)
		if err != nil {
			return cluster
		}

		obj := &unstructured.Unstructured{}
		err = yaml.Unmarshal(b, obj)
		if err != nil {
			return cluster
		}
		return obj.GetAnnotations()[ClusterAnnotation]
	}
	return cluster
}

// 针对table类型  加入集群标志
func addCustomColumn(obj *unstructured.Unstructured, req *http.Request) {
	//tb := v1.Table{}
	if obj.GetKind() == "Table" {
		if cd, ok := obj.Object["columnDefinitions"].([]interface{}); ok {
			cd = append(cd, map[string]interface{}{
				"name":        "Cluster",
				"description": "集群",
				"format":      "cluster",
				"type":        "string",
				"priority":    0,
			})
			//fmt.Println("cd len:" + strconv.Itoa(len(cd)))
			obj.Object["columnDefinitions"] = cd
		}

		if rows, ok := obj.Object["rows"].([]interface{}); ok {
			newRows := []interface{}{}
			for _, row := range rows {
				r := row.(map[string]interface{})
				if cells, ok := r["cells"].([]interface{}); ok {
					cells = append(cells, req.Header.Get("from_cluster"))
					row.(map[string]interface{})["cells"] = cells
					//fmt.Println("cells len:" + strconv.Itoa(len(cells)))
				}
				newRows = append(newRows, row)

			}

			obj.Object["rows"] = newRows
		}
	}
}

// 临时处理， 后面还要 大改
// 目前 只处理 Table 内容
// len(cnts) 一定>0 否则不会 调用此函数
func mergeResponse(cnts ...[]byte) ([]byte, error) {
	var ret *metav1.Table

	for _, cnt := range cnts {
		tmp := &unstructured.Unstructured{}
		if err := tmp.UnmarshalJSON(cnt); err == nil {
			if tmp.GetKind() == "Table" { // 后面要改。 因为这个类型 不支持 client-go
				tb := &metav1.Table{}
				err = runtime.DefaultUnstructuredConverter.
					FromUnstructured(tmp.Object, tb)
				if err != nil {
					continue
				}
				if ret == nil {
					ret = tb
				} else {
					ret.Rows = append(ret.Rows, tb.Rows...)
				}
			}
		} else {
			//	log.Println("错误是", string(cnt), "aaaa", len(cnt))
		}
	}
	if ret == nil {
		return cnts[0], nil //  临时处理。 代表没有解析到， 临时返回第一个
	} else {
		return json.Marshal(ret)
	}

}

// 克隆 请求对象 。 这一步 必须要在serveHttp 执行之前执行
func cloneRequest(srcReq *http.Request, cluster string) *http.Request {
	cloneRequest := srcReq.Clone(srcReq.Context())
	if srcReq.Body != nil {
		body, err := ioutil.ReadAll(srcReq.Body)
		if err != nil {
			log.Fatalln(err)
		}
		srcReq.Body = ioutil.NopCloser(bytes.NewReader(body))
		cloneRequest.Body = ioutil.NopCloser(bytes.NewReader(body))
	}
	cloneRequest.Header.Set("from_cluster", cluster)
	fmt.Println(cloneRequest.Header)
	return cloneRequest
}

// 用户多集群 输出合并后的结果
func writeResponse(writer http.ResponseWriter, src *MyWriter) {
	writer.Header().Set("Content-type", "application/json")
	if len(src.content) == 1 {
		writer.Header().Set("Content-Length", strconv.Itoa(len(src.content[0])))
		writer.Write(src.content[0])
	} else if len(src.content) > 1 {
		b, err := mergeResponse(src.content...)
		if err != nil {
			return
		}
		writer.Header().Set("Content-Length", strconv.Itoa(len(b)))
		writer.Write(b)
	}
}
