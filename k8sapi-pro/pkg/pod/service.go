package pod

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type PodSvc struct {
	Helper *PageHelper `inject:"-"`
	Db     *gorm.DB    `inject:"-"`
}

func NewPodSvc() *PodSvc {
	return &PodSvc{}
}

// 分页PODS的输出
func (this *PodSvc) PagePods(cluster, ns string, page, size int) *ItemsPage {
	pods := this.ListByNs(cluster, ns).([]*Pod)
	readyCount := 0 //就绪的pod数量
	allCount := 0   //总数量
	ipods := make([]interface{}, len(pods))
	for i, pod := range pods {
		allCount++
		ipods[i] = pod
		if pod.IsReady {
			readyCount++
		}
	}
	return this.Helper.PageResource(
		page,
		size,
		ipods).SetExt(gin.H{"ReadyNum": readyCount, "AllNum": allCount})
}

func (this *PodSvc) GetPodNum(cluster, nodeName string) (num int64) {
	if list, err := models.ListNoNamespaced(cluster, "Pod", this.Db); err == nil {
		for _, resource := range list {
			pod := utils.Convert([]byte(resource.Object)).(*v1.Pod)
			if pod.Spec.NodeName == nodeName {
				num++
			}
		}
	}
	return
}

func (this *PodSvc) ListByNs(cluster, ns string) interface{} {
	ret := make([]*Pod, 0)
	if list, err := models.List(cluster, ns, "Pod", this.Db); err == nil { //db查询
		for _, resource := range list {
			pod := utils.Convert([]byte(resource.Object)).(*v1.Pod) //转换对象
			ret = append(ret, &Pod{
				Name:      pod.Name,
				NameSpace: pod.Namespace,
				Images:    utils.GetImagesByPod(pod.Spec.Containers),
				NodeName:  pod.Spec.NodeName,
				Phase:     string(pod.Status.Phase), // 阶段
				IsReady:   utils.GetPodIsReady(pod), //是否就绪
				IP:        []string{pod.Status.PodIP, pod.Status.HostIP},
				// 定义好event的map和handler，并在config里注入
				// 后面根据所需要的规则，从pod的属性拼接一下从map中取即可
				Message:    this.getEventMessage(cluster, ns, pod.Name),
				CreateTime: pod.CreationTimestamp.Format("2006-01-02 15:04:05"),
			})
		}
	}
	return ret
}

func (this *PodSvc) getEventMessage(cluster, ns, podName string) string {
	name := fmt.Sprintf("Pod_%s", podName)
	r, err := models.Take(cluster, ns, name, "Event", this.Db)
	goft.Error(err)
	if r.Object != "" {
		o := &unstructured.Unstructured{}
		json.Unmarshal([]byte(r.Object), o)
		return o.Object["message"].(string)
	}
	return ""
}

func (this *PodSvc) GetPodContainer(cluster, ns, name string) []*ContainerModel {
	ret := make([]*ContainerModel, 0)
	r, err := models.Take(cluster, ns, name, "Pod", this.Db)
	pod := utils.Convert([]byte(r.Object)).(*v1.Pod) //转换对象
	goft.Error(err)
	for _, item := range pod.Spec.Containers {
		ret = append(ret, &ContainerModel{
			Name: item.Name,
		})
	}
	return ret
}
