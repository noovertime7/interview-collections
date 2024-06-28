package handlers

import (
	"github.com/octoboy233/mypodtrace/pkg/exporter"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	v1 "k8s.io/api/core/v1"
)

type EventHandlers struct {
	provider *trace.TracerProvider
}

func (e EventHandlers) OnAdd(obj interface{}, isInInitialList bool) {
	if event, ok := obj.(*v1.Event); ok && !isInInitialList {
		if event.InvolvedObject.Kind == "Pod" {
			if !IsTestResource(event.InvolvedObject.Name) {
				return
			}

			uid := event.InvolvedObject.UID
			//rootSpan := oteltrace.SpanFromContext(CtxMapper[uid].ctx)
			//if !rootSpan.IsRecording() {
			//	return //否则traces页面直接显示事件
			//}
			tracer := e.provider.Tracer("events")
			if info, ok := CtxMapper.Get(uid); ok {
				rootCtx := info.rootCtx
				_, span := tracer.Start(rootCtx, event.Reason)
				defer span.End()
				span.SetAttributes(
					attribute.String("message", event.Message),
					attribute.String("type", event.Type),
				)
			} else {
				//panic(fmt.Sprintf("can't find parent span for event:%s", event.InvolvedObject.Name))
			}
		}
	}
}

func (e EventHandlers) OnUpdate(oldObj, newObj interface{}) {
}

func (e EventHandlers) OnDelete(obj interface{}) {
}

func NewEventHandlers() *EventHandlers {
	return &EventHandlers{
		provider: exporter.NewJaegerProvider("event"),
	}
}
