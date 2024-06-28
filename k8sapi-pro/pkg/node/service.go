package node

import (
	"github.com/shenyisyn/goft-gin/goft"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/pkg/pod"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
)

type NodeSvc struct {
	Db      *gorm.DB        `inject:"-"`
	McliMap *models.McliMap `inject:"-"`
	PodSvc  *pod.PodSvc     `inject:"-"`
}

// 保存时用的
func (this *NodeSvc) LoadOrginNode(cluster, nodeName string) *v1.Node {
	if resource, err := models.Take(cluster, "", nodeName, "Node", this.Db); err == nil {
		node := utils.Convert([]byte(resource.Object)).(*v1.Node)
		return node
	} else {
		goft.Error(err)
	}
	return nil
}

// 加载node信息， 给编辑用的
func (this *NodeSvc) LoadNode(cluster, nodeName string) *NodeModel {
	node := this.LoadOrginNode(cluster, nodeName)
	return &NodeModel{
		Name:         node.Name,
		IP:           node.Status.Addresses[0].Address,
		HostName:     node.Status.Addresses[1].Address,
		OriginLabels: node.Labels,
		OriginTaints: node.Spec.Taints,
	}
}

// 显示所有节点
func (this *NodeSvc) ListAllNodes(cluster string) (ret []*NodeModel) {
	if resources, err := models.ListNoNamespaced(cluster, "Node", this.Db); err == nil {
		ret = make([]*NodeModel, len(resources))
		for i, resource := range resources {
			node := utils.Convert([]byte(resource.Object)).(*v1.Node)
			nodeUsage := utils.GetNodeUsage((*this.McliMap)[cluster], node)
			ret[i] = &NodeModel{
				Name:     node.Name,
				IP:       node.Status.Addresses[0].Address,
				HostName: node.Status.Addresses[1].Address,
				Labels:   utils.LabelsFilter(node.Labels),
				Taints:   utils.TaintsFilter(node.Spec.Taints),
				Capacity: &NodeCapacity{
					CPU:    node.Status.Capacity.Cpu().Value(),
					Memory: node.Status.Capacity.Memory().Value(),
					Pod:    node.Status.Capacity.Pods().Value(),
				},
				Usage: &NodeUsage{
					Pod:    this.PodSvc.GetPodNum(cluster, node.Name),
					CPU:    nodeUsage[0],
					Memory: nodeUsage[1],
				},
				CreateTime: node.CreationTimestamp.Format("2006-01-02 15:04:05"),
			}
		}
	}
	return
}
