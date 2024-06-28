package handlers

import (
	"github.com/octoboy233/mypodtrace/pkg/exporter"
	"go.opentelemetry.io/otel/sdk/trace"
	v1 "k8s.io/api/apps/v1"
)

type ReplicaSetHandlers struct {
	provider *trace.TracerProvider
}

func NewReplicaSetHandlers() *ReplicaSetHandlers {
	return &ReplicaSetHandlers{
		provider: exporter.NewJaegerProvider("replicaSet"),
	}
}

func (d ReplicaSetHandlers) OnAdd(obj interface{}, isInInitialList bool) {
	if rs, ok := obj.(*v1.ReplicaSet); ok && !isInInitialList {
		if !IsTestResource(rs.Name) {
			return
		}
		tracer := d.provider.Tracer("replicaSet")
		if info, ok := CtxMapper.Get(rs.OwnerReferences[0].UID); ok {
			parentCtx := info.ctx
			ctx, span := tracer.Start(parentCtx, rs.Name)
			defer span.End()
			CtxMapper.Add(rs.UID, spanInfo{
				ctx: ctx,
			})
		} else {
			//panic(fmt.Sprintf("can't find parent span for %s", rs.Name))
		}

	}
}

func (d ReplicaSetHandlers) OnUpdate(oldObj, newObj interface{}) {

}

func (d ReplicaSetHandlers) OnDelete(obj interface{}) {

}
