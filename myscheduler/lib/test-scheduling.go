package lib

import (
	"context"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	frameworkruntime "k8s.io/kubernetes/pkg/scheduler/framework/runtime"
	"time"
)

//出自/pkg/capacityscheduling 只留了主体框架，简化了大部分

const TestSchedulingName = "test-scheduling" //pod模版必须指定schedulerName，才可以被指定的调度插件加载

type TestScheduling struct {
	fac  informers.SharedInformerFactory
	args *Args
}

type Args struct {
	MaxPods int `json:"maxPods,omitempty"`
}

func (*TestScheduling) Name() string { //实现framework.Plugin的接口方法
	return TestSchedulingName
}

// keypoint 如果同时指定了nodeName和schedulerName，会使调度器完全失效。如果只有一个node，则会跳过打分阶段。
// interview可以讲个小坑 因为prefilter失效了，本来prefilter可以让新进来的pod pending
func NewTestScheduling(configuration runtime.Object, f framework.Handle) (framework.Plugin, error) {
	args := &Args{}
	if err := frameworkruntime.DecodeInto(configuration, args); err != nil { //由配置文件注入参数，并通过configuration获取
		return nil, err
	}
	return &TestScheduling{
		fac:  f.SharedInformerFactory(),
		args: args,
	}, nil //注入informer工厂
}

//以上是基本结构
//如果我们要实现不同接入点（pre-filter、filter、scoring）等，则需要为它实现接口方法，下面示例为pre-filter

// 这里用来检查并快速生成接口方法
var _ framework.PreFilterPlugin = &TestScheduling{}
var _ framework.FilterPlugin = &TestScheduling{}
var _ framework.ScorePlugin = &TestScheduling{}
var _ framework.PreScorePlugin = &TestScheduling{}
var _ framework.PermitPlugin = &TestScheduling{}

// 业务方法 也是prefilter类型的接口方法实现，指定prefilter的类型在插件pod的配置文件中
func (s *TestScheduling) PreFilter(ctx context.Context, state *framework.CycleState, p *v1.Pod) *framework.Status {
	klog.V(3).Infof("预过滤")
	pods, err := s.fac.Core().V1().Pods().Lister().Pods(p.Namespace).List(labels.Everything())
	if err != nil {
		return framework.NewStatus(framework.Error, err.Error())
	}
	if len(pods) > s.args.MaxPods { //通过参数获取，而不是写死
		//会体现在需要调度pod的事件中
		return framework.NewStatus(framework.Unschedulable, "pod数量超过了最大限制")
	}
	return framework.NewStatus(framework.Success)
}

// 这个方法是在生成pod或删除pod时产生一些需要评估的内容
func (s *TestScheduling) PreFilterExtensions() framework.PreFilterExtensions {
	return s //同样通过返回自身来快速生成下面两个方法
}

func (s *TestScheduling) AddPod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToAdd *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (s *TestScheduling) RemovePod(ctx context.Context, state *framework.CycleState, podToSchedule *v1.Pod, podInfoToRemove *framework.PodInfo, nodeInfo *framework.NodeInfo) *framework.Status {
	return nil
}

func (s *TestScheduling) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	klog.V(3).Infof("过滤节点")
	for k, v := range nodeInfo.Node().Labels {
		if k == "scheduling" && v != "true" {
			return framework.NewStatus(framework.Unschedulable, "设置了不可调度的标签")
		}
	}
	return framework.NewStatus(framework.Success)
}

type NodeMem struct {
	m map[string]float64 //节点内存名称--->内存空闲百分比
}

func (n *NodeMem) Clone() framework.StateData {
	return &NodeMem{m: n.m}
}

// 预打分 CycleState 主要负责调度流程中一些数据的保存，所有插件均可存取且内置了锁保证线程安全
func (s *TestScheduling) PreScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	nm := &NodeMem{
		m: make(map[string]float64),
	}
	nm.m["node-01"] = 0.4 //好比， 只有40%的空暇内存
	nm.m["node-02"] = 0.6
	state.Write("nodeMem", nm)
	klog.Info("预打分阶段：保存数据成功")
	return framework.NewStatus(framework.Success)
}

// 打出的分数区间是0-100
func (s *TestScheduling) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	getNodeMem, err := state.Read("nodeMem")
	if err != nil {
		return 0, framework.NewStatus(framework.Unschedulable)
	}
	max := 50.0
	per, ok := getNodeMem.(*NodeMem).m[nodeName]
	if ok {
		klog.Infof("打分阶段，%s得分是:%d", nodeName, int64(max*per))
		return int64(max * per), framework.NewStatus(framework.Success)
	} else {
		return 5, framework.NewStatus(framework.Success)
	}
}

// 因为内置了多个打分的插件且分数超过100会报错，需要使用归一方法，使最后得分落在0～100分
// 这里用了最简单的归一算法 （x-min）/max-min *100 这样一定会落在[min,max]
func (s *TestScheduling) ScoreExtensions() framework.ScoreExtensions {
	return s
}
func (s *TestScheduling) NormalizeScore(ctx context.Context, state *framework.CycleState, p *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	var min, max int64 = 0, 0
	//先求出 最小分数和最大分数
	for _, score := range scores {
		if score.Score < min {
			min = score.Score
		}
		if score.Score > max {
			max = score.Score
		}
	}
	if max == min {
		min = min - 1
	}

	for i, score := range scores {
		scores[i].Score = (score.Score - min) * framework.MaxNodeScore / (max - min)
		klog.Infof("节点: %v, Score: %v   Pod:  %v", scores[i].Name, scores[i].Score, p.GetName())
	}
	return framework.NewStatus(framework.Success, "")
}

// schedule的最后一道，判定条件，失败可以delay 然后重新入列
func (s *TestScheduling) Permit(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	_, err := s.fac.Core().V1().Pods().Lister().Pods("default").
		Get("nginx-66b6c48dd5-t7jph")
	if err != nil {
		klog.Info("Permit阶段：前置POD没有，需要等待10秒")
		return framework.NewStatus(framework.Wait), time.Second * 10
	} else {
		klog.Info("Permit阶段：通过，进入绑定周期")
		return framework.NewStatus(framework.Success), 0
	}
}
