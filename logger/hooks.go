package lggr

import (
	"context"
	"go.uber.org/zap"
)

var hooks []LoggerHook

type LoggerHook interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Debug(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Panic(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
	Critical(ctx context.Context, msg string, fields ...zap.Field)
	CriticalPanic(ctx context.Context, msg string, fields ...zap.Field)
	CriticalFatal(ctx context.Context, msg string, fields ...zap.Field)
}

func AddHook(hook LoggerHook) {
	hooks = append(hooks, hook)
}
