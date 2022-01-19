package fct_temporal

import (
	"context"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"go.opentelemetry.io/otel"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	tp_otel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/workflow"
	"os"
	"time"
)

var temporalClient client.Client

type TemporalSetupConfig struct {
	Namespace            string
	NamespaceDescription string
	Endpoint             string
}

func SetupTemporal(config *TemporalSetupConfig) client.Client {
	var logg TemporalZapLogger

	temporalLogger := TemporalZapLogger{logger: lggr.Get("temporal-internal").Logger}

	// First, ensure the desired namespace exists
	nsc, err := client.NewNamespaceClient(attachTracer(client.Options{
		HostPort: os.Getenv("TEMPORAL_HOST_PORT"),
		Logger:   logg,
	}))
	if err != nil {
		panic(err)
	}

	_, err = nsc.Describe(context.Background(), config.Namespace)
	if err != nil {
		if _, ok := err.(*serviceerror.NotFound); ok {
			// Need to create namespace
			err = nsc.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
				Namespace:                        config.Namespace,
				Description:                      config.NamespaceDescription,
				WorkflowExecutionRetentionPeriod: DurPtr(24 * 7 * time.Hour), // Save workflow execution logs for 1 week
				IsGlobalNamespace:                false,
			})
			if err != nil {
				panic(err)
			}

			// Poll for workspace creation
			for {
				_, err = nsc.Describe(context.Background(), config.Namespace)
				if err != nil {
					if _, ok := err.(*serviceerror.NotFound); ok {
						// Wait after namespace registration to give temporal a chance to catch up
						time.Sleep(1 * time.Second)
						continue
					} else {
						panic(err)
					}
				}

				break
			}
		} else {
			panic(err)
		}
	}

	// Close the namespace client, it is no longer needed
	nsc.Close()

	c, err := client.NewClient(attachTracer(client.Options{
		HostPort:           config.Endpoint,
		Namespace:          config.Namespace,
		ContextPropagators: []workflow.ContextPropagator{NewContextPropagator()},
		Logger:             temporalLogger,
	}))
	if err != nil {
		panic(err)
	}

	temporalClient = c

	// TODO figure out how to wait for namespace creation

	return temporalClient
}

func attachTracer(opts client.Options) client.Options {
	tracingInterceptor, err := tp_otel.NewTracingInterceptor(tp_otel.TracerOptions{
		Tracer:            otel.Tracer("temporal"),
		TextMapPropagator: otel.GetTextMapPropagator(),
	})
	if err != nil {
		panic(err)
	}

	opts.Interceptors = append(opts.Interceptors, tracingInterceptor)

	return opts
}

func DurPtr(d time.Duration) *time.Duration {
	return &d
}
