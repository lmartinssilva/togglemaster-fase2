package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// initTelemetry configura o OpenTelemetry (traces via OTLP/HTTP).
//
// Só exporta quando OTEL_EXPORTER_OTLP_ENDPOINT está definida (no cluster, via
// ConfigMap). Localmente e no CI fica desligado. O propagador W3C TraceContext
// é sempre registrado: é ele que costura o trace entre os serviços.
func initTelemetry(serviceName string) func(context.Context) error {
	noop := func(context.Context) error { return nil }

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") == "" {
		log.Println("OpenTelemetry desativado (OTEL_EXPORTER_OTLP_ENDPOINT não definida)")
		return noop
	}

	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx) // lê OTEL_EXPORTER_OTLP_* do ambiente
	if err != nil {
		log.Printf("Falha ao criar exporter OTLP; seguindo sem tracing: %v", err)
		return noop
	}

	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes("", attribute.String("service.name", serviceName)))
	if err != nil {
		log.Printf("Falha ao criar resource OTel: %v", err)
		return noop
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(2*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	log.Printf("OpenTelemetry ativo: service=%s", serviceName)
	return tp.Shutdown
}

// instrumentHandler envolve o mux: cada request vira um span SERVER
// ("GET /evaluate"). O /health é ignorado (ruído das probes do kubelet).
func instrumentHandler(h http.Handler) http.Handler {
	return otelhttp.NewHandler(h, "http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
		otelhttp.WithFilter(func(r *http.Request) bool { return r.URL.Path != "/health" }),
	)
}

// startClientSpan abre um span CLIENT para uma dependência externa
// (Redis, Postgres, SQS). O atributo peer.service faz a dependência aparecer
// como um nó no Service Map do Datadog.
func startClientSpan(ctx context.Context, name, peer string) (context.Context, trace.Span) {
	return otel.Tracer("togglemaster").Start(ctx, name,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.String("peer.service", peer)))
}

// endSpan fecha o span registrando o erro (se houver).
func endSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}
