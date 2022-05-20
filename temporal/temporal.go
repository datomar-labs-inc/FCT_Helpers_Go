package fcttemporal

import (
	"context"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/friendsofgo/errors"
	"go.opentelemetry.io/otel"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	tp_otel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"strings"
	"time"
)

var temporalClient client.Client

type TemporalSetupConfig struct {
	Namespace            string
	NamespaceDescription string
	Endpoint             string
	ConnectRetries       int
}

func SetupTemporal(config *TemporalSetupConfig) client.Client {
	tries := 0

	for {
		if tries > config.ConnectRetries {
			panic("failed to connect to temporal")
		}

		temporalClient, err := setupTemporal(config)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
				lggr.Get("setup-temporal").Error("could not connect to temporal, retrying...", zap.Error(err))
				time.Sleep(10 * time.Second)
				tries++
				continue
			}

			panic(err)
		}

		return temporalClient
	}
}

func setupTemporal(config *TemporalSetupConfig) (client.Client, error) {
	var logg TemporalZapLogger

	temporalLogger := TemporalZapLogger{logger: lggr.Get("temporal-internal").Logger.WithOptions(zap.AddCallerSkip(1))}

	// First, ensure the desired namespace exists
	nsc, err := client.NewNamespaceClient(attachTracer(client.Options{
		HostPort: config.Endpoint,
		Logger:   logg,
	}))
	if err != nil {
		return nil, err
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
				return nil, err
			}

			// Poll for workspace creation
			for {
				_, err = nsc.Describe(context.Background(), config.Namespace)
				if err != nil {
					if _, ok := err.(*serviceerror.NotFound); ok {
						// Wait after namespace registration to give temporal a chance to catch up
						time.Sleep(1 * time.Second)
						continue
					}

					return nil, err
				}

				break
			}
		} else {
			return nil, err
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

	return temporalClient, nil
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
