package main

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/mycore"
	"k8s.io/kubernetes/pkg/node"
	"log"
	"path/filepath"
	"time"
)

func GetCli() (*kubernetes.Clientset, error) {
	file := "e-config"
	configPath := filepath.Join(clientcmd.RecommendedConfigDir, file)
	cfg, err := clientcmd.BuildConfigFromFlags("", configPath) //可以用kubelet.config
	if err != nil {
		return nil, err
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func main() {
	klog.InitFlags(nil)

	cli, _ := GetCli()
	nodeName := "mynode"

	go func() {
		//先更新节点状态，然后不断续租
		err := node.UpdateNodeStatus(cli, nodeName)
		if err != nil {
			log.Fatal(err)
		}
		//hang
		node.StartLeaseController(cli, nodeName)
	}()

	k := mycore.NewMyKubelet(cli, nodeName)

	k.SetOnAdd(func(ctx *mycore.CallBackContent) error {
		fmt.Println("OnAdd，" + ctx.Pod.Name)
		ctx.AddEvent("OnAddCallBack", "OnAddCallBack")

		//设置成running状态
		pod_status := mycore.GetPodStatusReady(ctx.Pod)
		k.SetCache(ctx.Pod.UID, pod_status)

		//业务函数 边缘任务 mysql查表 之类的。。
		cmds := ctx.GetPodCommand()
		for _, cmd := range cmds {
			//执行命令行
			//cmd.Run() //cmd = "exit 1"
			//设置状态
			//ctx.SetContainerExited(cmd.ContainerName, cmd.ExitCode) // pod status会显示error

			fmt.Println(cmd)
		}

		time.Sleep(time.Second * 3)
		//执行完成 退出
		//似乎设置sandbox状态为notReady后再更新状态也无效了，说明容器不停重启sandbox不会重启
		ctx.SetPodCompleted()

		return nil
	})

	k.Run()

}
