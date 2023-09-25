package fcttemporal

import (
	"context"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	lggr "github.com/datomar-labs-inc/FCT_Helpers_Go/logger"
	"github.com/friendsofgo/errors"
	"go.opentelemetry.io/otel"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	tp_otel "go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strings"
	"time"
)

var temporalClient client.Client

type TemporalSetupConfig struct {
	Namespace             string
	NamespaceDescription  string
	Endpoint              string
	SkipNamespaceCreation bool
	ConnectRetries        int
}

func SetupTemporal(config *TemporalSetupConfig, logger *lggr.LogWrapper) client.Client {
	tries := 0

	for {
		if tries > config.ConnectRetries {
			panic("failed to connect to temporal")
		}

		temporalClient, err := setupTemporalInternal(config, logger)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded") {
				logger.Error("could not connect to temporal, retrying...", zap.Error(err))
				time.Sleep(10 * time.Second)
				tries++
				continue
			}

			if !strings.Contains(err.Error(), "not found") {
				logger.Error(fmt.Sprintf("Experienced temporal error during setup, panicking: %s", err))
				panic(err)
			}
		}

		return temporalClient
	}
}

//revive:disable:cyclomatic This is fine
func setupTemporalInternal(config *TemporalSetupConfig, logger *lggr.LogWrapper) (client.Client, error) {
	temporalLogger := TemporalZapLogger{logger: logger.WithOptions(zap.AddCallerSkip(4)).With(zap.Namespace("@temporal"))}

	if !config.SkipNamespaceCreation {
		err := createNamespace(config, &temporalLogger)
		if err != nil {
			return nil, ferr.Wrap(err)
		}
	}

	c, err := client.Dial(client.Options{
		HostPort:           config.Endpoint,
		Namespace:          config.Namespace,
		ContextPropagators: []workflow.ContextPropagator{},
		Logger:             temporalLogger,
		ConnectionOptions: client.ConnectionOptions{
			DialOptions: []grpc.DialOption{grpc.WithTimeout(10 * time.Second)},
		},
	})
	if err != nil {
		logger.Error("could not connect to temporal via dial, panicking", zap.Error(err))
		panic(err)
	}

	temporalClient = c

	return temporalClient, nil
}

func createNamespace(config *TemporalSetupConfig, temporalLogger *TemporalZapLogger) error {
	// First, ensure the desired namespace exists
	namespaceClient, err := client.NewNamespaceClient(attachTracer(client.Options{
		HostPort: config.Endpoint,
		Logger:   temporalLogger,
	}))
	if err != nil {
		return ferr.Wrap(err)
	}

	_, err = namespaceClient.Describe(context.Background(), config.Namespace)
	if err != nil {
		namespaceNotFound := false

		if _, ok := err.(*serviceerror.NotFound); ok {
			namespaceNotFound = true
		}

		if _, ok := err.(*serviceerror.NamespaceNotFound); ok {
			namespaceNotFound = true
		}

		if namespaceNotFound {
			// Need to create namespace
			err = namespaceClient.Register(context.Background(), &workflowservice.RegisterNamespaceRequest{
				Namespace:                        config.Namespace,
				Description:                      config.NamespaceDescription,
				WorkflowExecutionRetentionPeriod: DurPtr(24 * 7 * time.Hour), // Save workflow execution logs for 1 week
				IsGlobalNamespace:                false,
			})
			if err != nil {
				return ferr.Wrap(err)
			}

			temporalLogger.Info("Waiting after initial temporal namespace creation")
			time.Sleep(30 * time.Second)

			// Poll for workspace creation
			for {
				_, err = namespaceClient.Describe(context.Background(), config.Namespace)
				if err != nil {
					if _, ok := err.(*serviceerror.NotFound); ok {
						// Wait after namespace registration to give temporal a chance to catch up
						time.Sleep(1 * time.Second)
						continue
					}

					if _, ok := err.(*serviceerror.NamespaceNotFound); ok {
						// Wait after namespace registration to give temporal a chance to catch up
						time.Sleep(1 * time.Second)
						continue
					}

					return ferr.Wrap(err)
				}

				break
			}
		} else {
			return ferr.Wrap(err)
		}
	}

	// Close the namespace client, it is no longer needed
	namespaceClient.Close()

	return nil
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
