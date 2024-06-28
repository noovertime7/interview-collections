package websocket

import (
	"github.com/gin-gonic/gin"
	"github.com/shenyisyn/goft-gin/goft"
	"k8s.io/client-go/tools/remotecommand"
	"k8sapi-pro/pkg/websocket/wscore"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
	"log"
)

type WsCtl struct {
	RestConfigMap *models.RestConfigMap `inject:"-"`
	CliMap        *models.CliMap        `inject:"-"`
	SysConfig     *models.SysConfig     `inject:"-"`
}

func (this *WsCtl) Build(goft *goft.Goft) {
	goft.Handle("GET", "/ws", this.Connect)
	goft.Handle("GET", "/v1/:cluster/podws", this.PodConnect)
	goft.Handle("GET", "/nodews", this.NodeConnect)
}

func (this *WsCtl) Connect(c *gin.Context) string {
	client, err := wscore.Upgrader.Upgrade(c.Writer, c.Request, nil) //升级成websocket
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	wscore.ClientMap.Store(client)
	return "true"
}

func (this *WsCtl) PodConnect(c *gin.Context) (v goft.Void) {
	//获取容器相关对应参数
	cluster := c.Param("cluster")
	ns := c.Query("ns")
	pod := c.Query("name")
	container := c.Query("cname")
	//升级http客户端到ws客户端
	wsClient, err := wscore.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	shellClient := wscore.NewWsShellClient(wsClient) //以ws客户端构建实现reader/writer接口的对象
	//实现websocket接口方法，就是使其有资格成为一个writer。reader。
	//真正的write，read过程不关注，这个是自实现的
	//然后将其作为标准输入输出
	//真正的write。read过程就是，往websocket客户端读写数据，然后前端就会接收到
	err = utils.HandleCommand(ns, pod, container, (*this.CliMap)[cluster], (*this.RestConfigMap)[cluster], []string{"sh"}).
		Stream(remotecommand.StreamOptions{ //以流的方式来读取结果
			Stdin:  shellClient,
			Stdout: shellClient,
			Stderr: shellClient,
			Tty:    true,
		})
	goft.Error(err)
	return
}

func (this *WsCtl) NodeConnect(c *gin.Context) (v goft.Void) {
	nodeName := c.Query("node")
	nodeConfig := utils.GetNodeConfig(this.SysConfig, nodeName) //读取配置文件
	wsClient, err := wscore.Upgrader.Upgrade(c.Writer, c.Request, nil)
	goft.Error(err)
	shellClient := wscore.NewWsShellClient(wsClient)
	//session, err := helpers.SSHConnect(helpers.TempSSHUser, helpers.TempSSHPWD, helpers.TempSSHIP, 22)
	session, err := utils.SSHConnect(nodeConfig.User, nodeConfig.Pass, nodeConfig.Ip, 22)
	goft.Error(err)
	defer session.Close()
	session.Stdout = shellClient
	session.Stderr = shellClient
	session.Stdin = shellClient
	err = session.RequestPty("xterm-256color", 300, 500, utils.NodeShellModes)
	goft.Error(err)

	err = session.Run("bash")
	goft.Error(err)
	return
}

func (this *WsCtl) Name() string {
	return "WsCtl"
}

func NewWsCtl() *WsCtl {
	return &WsCtl{}
}
