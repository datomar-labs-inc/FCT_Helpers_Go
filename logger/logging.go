package lggr

import (
	"context"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/logger/prettylogger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"os"
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
	ec := prettylogger.NewEncoderConfig()
	ec.CallerKey = "caller"

	enc := prettylogger.NewEncoder(ec)

	return zap.New(zapcore.NewCore(
		enc,
		os.Stdout,
		zapcore.DebugLevel,
	)).With(zap.Namespace("@payload"))
}

// NewDevTool only logs warnings/errors
func NewDevTool() *LogWrapper {
	ec := prettylogger.NewEncoderConfig()
	ec.CallerKey = "caller"

	enc := prettylogger.NewEncoder(ec)

	return zap.New(zapcore.NewCore(
		enc,
		os.Stdout,
		zapcore.WarnLevel,
	)).With(zap.Namespace("@payload"))
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
