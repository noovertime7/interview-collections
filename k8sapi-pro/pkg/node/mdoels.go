package node

import v1 "k8s.io/api/core/v1"

// 容量
type NodeCapacity struct {
	CPU    int64
	Memory int64
	Pod    int64
}

// 使用率/使用情况
type NodeUsage struct {
	CPU    float64
	Memory float64
	Pod    int64
}

// 保存用
type PostNodeModel struct {
	Name         string
	OriginLabels map[string]string //原始标签 ---->前端 是一个对象
	OriginTaints []v1.Taint        //原始污点
}

// 节点模型
type NodeModel struct {
	Name         string
	IP           string
	HostName     string
	CreateTime   string
	OriginLabels map[string]string //原始标签 用于获取到后编辑保存用
	OriginTaints []v1.Taint        //原始污点
	Labels       []string          //用于列表展现 返回节点对象数据时会过滤一些原生的标签
	Taints       []string
	Capacity     *NodeCapacity
	Usage        *NodeUsage
}
