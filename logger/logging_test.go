package lggr

import (
	"go.uber.org/zap"
	"testing"
)

func TestGetAction(t *testing.T) {
	Get("some-action").
		Category(CategoryHost).
		Info("This is a test message", zap.Int("number", 69))

	Get("some-action").Span(nil).Info("something")
}
