package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"github.com/ServiceWeaver/weaver-kube/tool"
)

const jaegerPort = 14268

func main() {
	// Ajuste para o nome do service Jaeger no seu cluster.
	// Use `kubectl get svc` e veja se é "jaeger-collector" ou só "jaeger".
	jaegerURL := fmt.Sprintf("http://jaeger:%d/api/traces", jaegerPort)

	endpoint := jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerURL))
	traceExporter, err := jaeger.New(endpoint)
	if err != nil {
		panic(err)
	}

	handleTraceSpans := func(ctx context.Context, spans []trace.ReadOnlySpan) error {
		return traceExporter.ExportSpans(ctx, spans)
	}

	tool.Run("customkube", tool.Plugins{
		HandleTraceSpans: handleTraceSpans,
	})
}
