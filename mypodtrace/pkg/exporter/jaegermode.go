package exporter

import (
	"github.com/octoboy233/mypodtrace/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
)

const jaeger_server = "http://127.0.0.1:14268/api/traces"

// 资源 ： 可观测实体
func NewJaegerResource(resourceName string) *resource.Resource {
	defaultRes := resource.Default()

	r, err := resource.Merge(
		defaultRes,
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(resourceName),
			semconv.ServiceVersionKey.String(config.GetClusterName()+"@"+config.GetClusterVersion()),
		),
	)
	if err != nil {
		log.Fatalln(err)
	}
	return r
}

// 定义导出器
func NewJaegerExporter() (trace.SpanExporter, error) {
	return jaeger.New(
		//服务端起了一个all-in-one的jaeger容器，并展示webUI
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaeger_server)),
	)
}

// 创建 provider
func NewJaegerProvider(resourceName string) *trace.TracerProvider {
	exporter, err := NewJaegerExporter()
	if err != nil {
		log.Fatalln(err)
	}
	res := NewJaegerResource(resourceName)

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp

}
