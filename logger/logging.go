package lggr

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type ContextKeyType string

const ContextKey = ContextKeyType("__lggr.log_wrapper")

//type LogWrapper struct {
//	CallerSkip     int         `json:"caller_skip"`
//	DetachedFields []zap.Field `json:"detached_fields"`
//	log            *zap.Logger
//	ctx            context.Context
//}

type LogWrapper = zap.Logger

func New() *LogWrapper {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}

	lg, err := config.Build()
	if err != nil {
		panic(err)
	}

	return lg.With(zap.Namespace("@payload"))
}

func NewDev() *LogWrapper {
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return lg
}

func NewTest(t *testing.T) *LogWrapper {
	testLogger := zaptest.NewLogger(t)

	return testLogger
}

func FromContext(ctx context.Context) *LogWrapper {
	if ctx == nil {
		return nil
	}

	if lggr, ok := ctx.Value(ContextKey).(*LogWrapper); ok {
		return lggr
	}

	return nil
}

func AttachToContext(parent context.Context, logger *LogWrapper) context.Context {
	if logger == nil {
		panic("missing logger")
	}

	return context.WithValue(parent, ContextKey, logger)
}
