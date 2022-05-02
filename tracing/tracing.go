package fcttracing

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"os"
)

func SetupTracing(serviceName string, sampler trace.Sampler) (closer func() error, err error) {
	return setupOpenTelemetry(context.Background(), serviceName, sampler)
}

func setupOpenTelemetry(ctx context.Context, serviceName string, sampler trace.Sampler) (func() error, error) {
	res, err := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceNameKey.String(serviceName),
	))
	if err != nil {
		return nil, err
	}

	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(os.Getenv("TRACES_ENDPOINT")),
		otlptracegrpc.WithInsecure(),
	)

	traceExporter, err := otlptrace.New(ctx, client)
	if err != nil {
		return nil, err
	}

	batchSpanProcessor := trace.NewBatchSpanProcessor(traceExporter)

	traceProvider := trace.NewTracerProvider(
		trace.WithSampler(sampler),
		trace.WithResource(res),
		trace.WithSpanProcessor(batchSpanProcessor),
	)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Return a shutdown handler
	return func() error {
		return traceProvider.Shutdown(ctx)
	}, nil
}
