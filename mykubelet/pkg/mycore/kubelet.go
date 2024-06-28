package mycore

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	kubecontainer "k8s.io/kubernetes/pkg/kubelet/container"
	kubetypes "k8s.io/kubernetes/pkg/kubelet/types"
	"os/exec"
	"time"
)

type MyKubelet struct {
	podCache                            *PodCache
	OnAdd, OnUpdate, OnDelete, OnRemove CallBackFunc //回调函数 用来支持一些指标 事件等
}

func (kl *MyKubelet) SetCache(uid types.UID, status *kubecontainer.PodStatus) {
	kl.podCache.InnerCache.Set(uid, status, nil, time.Now())
}

// 用于回调记录事件的结构体，在回调时包装传入
type CallBackContent struct {
	Pod      *v1.Pod
	recorder record.EventRecorder
	PodCache *PodCache
}

// 添加事件
func (c *CallBackContent) AddEvent(msg, reason string) {
	c.recorder.Event(c.Pod, v1.EventTypeNormal, reason, msg)
}

// 设置pod状态为completed
func (c *CallBackContent) SetPodCompleted() {
	status := GetPodStatusCompleted(c.Pod)
	c.PodCache.InnerCache.Set(c.Pod.UID, status, nil, time.Now())
}

// 获得pod的命令
// 这里用了struct来包装exec.Cmd,因为执行完要取对应container名称去设置状态
func (ctx *CallBackContent) GetPodCommand() (res []*ContainerCmd) {
	for _, c := range ctx.Pod.Spec.Containers {
		if len(c.Command) == 0 {
			continue
		}
		args := []string{}
		if len(c.Command) > 1 {
			args = c.Command[1:]
		}
		args = append(args, c.Args...)
		cmd := exec.Command(c.Command[0], args...)
		res = append(res, &ContainerCmd{
			ContainerName: c.Name,
			Cmd:           cmd,
		})
	}
	return
}

func (ctx *CallBackContent) SetContainerExited(containerName string, exitCode int) {
	status := GetPodStatusExited(ctx.Pod, containerName, exitCode)
	ctx.PodCache.InnerCache.Set(ctx.Pod.UID, status, nil, time.Now())
}

type CallBackFunc func(content *CallBackContent) error

func NewMyKubelet(cli *kubernetes.Clientset, nodeName string) *MyKubelet {
	return &MyKubelet{
		// 初始化了podConfig，用来 watch apiServer,监听pod的变化（在kubelet中是供syncLoop使用）
		podCache: NewPodCache(cli, nodeName),
	}
}

func (k *MyKubelet) SetOnAdd(onAdd CallBackFunc) {
	k.OnAdd = onAdd
}

func (k *MyKubelet) SetOnUpdate(onUpdate CallBackFunc) {
	k.OnUpdate = onUpdate
}

func (k *MyKubelet) SetOnRemove(onRemove CallBackFunc) {
	k.OnRemove = onRemove
}

func (k *MyKubelet) SetOnDelete(onDelete CallBackFunc) {
	k.OnDelete = onDelete
}

func (k *MyKubelet) Run() {
	klog.Info("边缘kubelet开始启动...")
	//模拟了syncLoop中对configCh的读取
	pc := k.podCache
	for event := range pc.PodConfig.Updates() { //相当于syncLoop 不断读Channel
		switch event.Op {
		case kubetypes.ADD:
			//这里（可以）添加将pod置为Ready并加入到podCache的逻辑，因此会自动同步到apiserver并显示成Running
			HandlerAddPod(event.Pods, pc, k.OnAdd)
			break
		case kubetypes.UPDATE:
			HandlerUpdatePod(event.Pods, pc, k.OnUpdate)
			break
		case kubetypes.REMOVE: //停止之前创建的podWorker
			HandlerRemovePod(event.Pods, pc, k.OnRemove)
			break
		case kubetypes.DELETE: //并非真的删除了，其实是修改了状态，删除是syncPod做的
			HandlerUpdatePod(event.Pods, pc, k.OnDelete)
			break
		default:
			break
		}
	}
}
