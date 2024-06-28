package node

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8sapi-pro/src/models"
	"log"
)

type NodeCtl struct {
	Svc    *NodeSvc       `inject:"-"`
	CliMap *models.CliMap `inject:"-"`
}

func (this *NodeCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/nodes", this.ListAll)
	goft.Handle("GET", "/:cluster/nodes/:node", this.LoadDetail) //加载详细
	goft.Handle("POST", "/:cluster/nodes", this.SaveNode)        //保存
}

func (this *NodeCtl) ListAll(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.Svc.ListAllNodes(cluster),
	}

}

func (this *NodeCtl) LoadDetail(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	nodeName := c.Param("node")
	return gin.H{
		"code": 20000,
		"data": this.Svc.LoadNode(cluster, nodeName),
	}

}

// 保存node
func (this *NodeCtl) SaveNode(c *gin.Context) goft.Json {
	cluster := c.Param("cluster")
	nodeModel := &PostNodeModel{}
	_ = c.ShouldBindJSON(nodeModel)
	node := this.Svc.LoadOrginNode(cluster, nodeModel.Name) //取出原始node 信息
	if node == nil {
		log.Fatal("no such node")
	}
	node.Labels = nodeModel.OriginLabels      //覆盖标签
	node.Spec.Taints = nodeModel.OriginTaints //覆盖 污点
	_, err := (*this.CliMap)[cluster].CoreV1().Nodes().Update(context.Background(), node, v1.UpdateOptions{})
	goft.Error(err)
	return gin.H{
		"code": 20000,
		"data": "success",
	}
}

func (this *NodeCtl) Name() string {
	return "NodeCtl"
}

func NewNodeCtl() *NodeCtl {
	return &NodeCtl{}
}
