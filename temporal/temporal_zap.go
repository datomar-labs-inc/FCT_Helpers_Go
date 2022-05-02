package fcttemporal

import (
	"fmt"
	"go.uber.org/zap"
)

type TemporalZapLogger struct {
	logger *zap.Logger
}

func (t TemporalZapLogger) Debug(msg string, keyvals ...any) {
	if t.logger == nil {
		return
	}

	t.logger.Debug(msg, CollectLogFields(keyvals)...)
}

func (t TemporalZapLogger) Info(msg string, keyvals ...any) {
	if t.logger == nil {
		return
	}

	t.logger.Info(msg, CollectLogFields(keyvals)...)
}

func (t TemporalZapLogger) Warn(msg string, keyvals ...any) {
	if t.logger == nil {
		return
	}

	t.logger.Warn(msg, CollectLogFields(keyvals)...)
}

func (t TemporalZapLogger) Error(msg string, keyvals ...any) {
	if t.logger == nil {
		return
	}

	t.logger.Error(msg, CollectLogFields(keyvals)...)
}

func CollectLogFields(keyvals []any) []zap.Field {
	var fields []zap.Field

	for i := 0; i < len(keyvals)/2; i++ {
		key := fmt.Sprintf("%v", keyvals[i*2])

		fields = append(fields, zap.Any(key, keyvals[i*2+1]))
	}

	return fields
}
