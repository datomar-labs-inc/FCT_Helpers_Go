package fcttracing

import (
	"context"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/go-logr/zapr"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"os"
)

func SetupTracing(serviceName string, logger *lggr.LogWrapper, sampler trace.Sampler) (closer func() error, err error) {
	return setupOpenTelemetry(context.Background(), serviceName, sampler, logger, nil)
}

func SetupTracingWithCustomSpanProcessor(serviceName string, logger *lggr.LogWrapper, sampler trace.Sampler, customSpanProcessor trace.SpanProcessor) (closer func() error, err error) {
	return setupOpenTelemetry(context.Background(), serviceName, sampler, logger, customSpanProcessor)
}

func setupOpenTelemetry(ctx context.Context, serviceName string, sampler trace.Sampler, logger *lggr.LogWrapper, customSpanProcessor trace.SpanProcessor) (func() error, error) {
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String(serviceName),
	))
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(os.Getenv("TRACES_ENDPOINT")),
		otlptracegrpc.WithInsecure(),
	)

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	if customSpanProcessor == nil {
		customSpanProcessor = trace.NewBatchSpanProcessor(traceExporter)
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(sampler),
		trace.WithResource(res),
		trace.WithSpanProcessor(customSpanProcessor),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otel.SetLogger(zapr.NewLogger(logger.GetInternalZapLogger()))

	// Return a shutdown handler
	return func() error {
		return traceProvider.Shutdown(ctx)
	}, nil
}
