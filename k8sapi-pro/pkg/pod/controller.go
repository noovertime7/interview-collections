package pod

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8sapi-pro/src/models"
	"net/http"
	"time"
)

type PodCtl struct {
	PodService *PodSvc        `inject:"-"`
	Helper     *PageHelper    `inject:"-"`
	CliMap     *models.CliMap `inject:"-"`
}

func (this *PodCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/:cluster/pods", this.GetAll)
	goft.Handle("GET", "/:cluster/pods/containers", this.Containers)
	goft.Handle("GET", "/:cluster/pods/logs", this.GetLogs)
}

func (this *PodCtl) GetAll(c *gin.Context) goft.Json {
	ns := c.DefaultQuery("ns", "default")
	page := c.DefaultQuery("current", "1") //当前页
	size := c.DefaultQuery("size", "5")
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.PodService.PagePods(cluster, ns,
			this.Helper.StrToInt(page, 1),
			this.Helper.StrToInt(size, 5)),
	}

}

// 获取 容器
func (this *PodCtl) Containers(c *gin.Context) goft.Json {
	ns := c.DefaultQuery("ns", "default")
	podname := c.DefaultQuery("name", "")
	cluster := c.Param("cluster")
	return gin.H{
		"code": 20000,
		"data": this.PodService.GetPodContainer(cluster, ns, podname),
	}

}

func (this *PodCtl) GetLogs(c *gin.Context) (v goft.Void) {
	cluster := c.Param("cluster")
	ns := c.DefaultQuery("ns", "default")
	podname := c.DefaultQuery("podname", "")
	cname := c.DefaultQuery("cname", "")
	req := (*this.CliMap)[cluster].CoreV1().Pods(ns).GetLogs(podname, &v1.PodLogOptions{
		Container: cname,
		Follow:    true,
	})
	//单次获取
	//res, err := req.DoRaw(context.Background())
	//流式获取
	//gin会给每个请求都起一个协程，不设超时时间就会阻塞在read
	cc, _ := context.WithTimeout(context.Background(), time.Minute*5)
	reader, err := req.Stream(cc)
	goft.Error(err)
	defer reader.Close()
	for {
		b := make([]byte, 1024)
		n, err := reader.Read(b)
		if err != nil && err == io.EOF {
			break
		}
		if n > 0 {
			//长连接分段传输
			c.Writer.Write(b[0:n])
			c.Writer.(http.Flusher).Flush()
		}
	}
	return
}

func (p PodCtl) Name() string {
	return "PodCtl"
}

func NewPodCtl() *PodCtl {
	return &PodCtl{}
}
