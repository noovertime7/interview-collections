package handlers

import (
	"fmt"
	"github.com/octoboy233/mypodtrace/pkg/exporter"
	"github.com/octoboy233/mypodtrace/pkg/k8ehelper"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
	v1 "k8s.io/api/core/v1"
)

//informer handler的处理函数

type PodHandlers struct {
	provider *trace.TracerProvider
}

func NewPodHandlers() *PodHandlers {
	return &PodHandlers{
		provider: exporter.NewJaegerProvider("pod"),
	}
}

func (p PodHandlers) OnAdd(obj interface{}, isInInitialList bool) {
	if pod, ok := obj.(*v1.Pod); ok && !isInInitialList { //必须是刚创建的pod
		if !IsTestResource(pod.Name) {
			return
		}

		// 创建一个新的Span
		tracer := p.provider.Tracer("pods")
		// 微服务请求链路要用propagation从请求头获取ctx，这里用缓存
		if info, ok := CtxMapper.Get(pod.OwnerReferences[0].UID); ok {
			parentCtx := info.ctx
			rootCtx, rootSpan := tracer.Start(parentCtx, pod.Name)
			ctx, span := tracer.Start(rootCtx, "pod lifecycle")
			defer span.End()
			defer rootSpan.End()
			//设置pod的节点和phase作为span的属性
			span.SetAttributes(
				attribute.KeyValue{
					Key:   "phase",
					Value: attribute.StringValue(string(pod.Status.Phase)),
				},
			)
			CtxMapper.Add(pod.UID, spanInfo{
				rootCtx: rootCtx,
				ctx:     ctx,
			})
		} else {
			//panic(fmt.Sprintf("can't find parent span for %s", pod.Name))
		}
	}
}

func (p PodHandlers) OnUpdate(oldObj, newObj interface{}) {
	if pod, ok := newObj.(*v1.Pod); ok {

		if !IsTestResource(pod.Name) {
			return
		}

		// 创建一个新的Span
		tracer := p.provider.Tracer("pods")

		//保证ctx与原pod一致，同一个pod的生命周期会聚合在一起
		if spanInfo, ok := CtxMapper.Get(pod.UID); ok {

			podInfo := k8ehelper.PrintPod(pod)
			//根据pod信息组装spanName
			spanName := fmt.Sprintf("%s(%s) - %s", pod.Name, podInfo.ContainerReady, podInfo.Reason)
			_, span := tracer.Start(spanInfo.ctx, spanName)
			defer span.End()

			//设置pod的节点和phase作为span的属性
			span.SetAttributes(
				attribute.KeyValue{
					Key:   "phase",
					Value: attribute.StringValue(string(pod.Status.Phase)),
				},
			)
		} else {
			//panic(fmt.Sprintf("can't find parent span for %s", pod.Name))
		}
	}
}

func (p PodHandlers) OnDelete(obj interface{}) {
	if pod, ok := obj.(*v1.Pod); ok {

		if !IsTestResource(pod.Name) {
			return
		}

		if info, ok := CtxMapper.Get(pod.UID); ok {
			rootSpan := oteltrace.SpanFromContext(info.rootCtx)
			if rootSpan.IsRecording() {
				spanName := fmt.Sprintf("%s: %s(%s)", pod.Spec.NodeName, pod.Name, "deleted")
				rootSpan.SetName(spanName)
				rootSpan.End()
			}
		} else {
			//panic(fmt.Sprintf("can't find parent span for %s", pod.Name))
		}
	}
}
