package handlers

import (
	"context"
	"fmt"
	"github.com/octoboy233/mypodtrace/pkg/exporter"
	"go.opentelemetry.io/otel/sdk/trace"
	v1 "k8s.io/api/apps/v1"
)

type DeploymentHandlers struct {
	provider *trace.TracerProvider
}

func NewDeploymentHandlers() *DeploymentHandlers {
	return &DeploymentHandlers{
		provider: exporter.NewJaegerProvider("deployment"),
	}
}

func (d DeploymentHandlers) OnAdd(obj interface{}, isInInitialList bool) {
	if deploy, ok := obj.(*v1.Deployment); ok && !isInInitialList {
		if !IsTestResource(deploy.Name) {
			return
		}
		tracer := d.provider.Tracer("deployment")
		spanName := fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)
		ctx, span := tracer.Start(context.Background(), spanName)
		defer span.End()
		CtxMapper.Add(deploy.UID, spanInfo{
			ctx: ctx,
		})
	}
}

func (d DeploymentHandlers) OnUpdate(oldObj, newObj interface{}) {

}

func (d DeploymentHandlers) OnDelete(obj interface{}) {

}
