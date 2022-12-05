package lggr

import (
	"context"
	"encoding/json"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

func NewContextPropagator(logger *LogWrapper) *ContextPropogator {
	return &ContextPropogator{
		logger: logger,
	}
}

type ContextPropogator struct {
	logger *LogWrapper
}

func (l *ContextPropogator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	lw := FromContext(ctx)

	if lw != nil {
		jsb, err := json.Marshal(lw)
		if err != nil {
			return err
		}

		pl, err := converter.GetDefaultDataConverter().ToPayload(jsb)
		if err != nil {
			return err
		}

		writer.Set(string(ContextKey), pl)
	}

	return nil
}

func (l *ContextPropogator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if pl, ok := reader.Get(string(ContextKey)); ok {
		loggerFromJSON, err := NewFromJSON(l.logger, pl.Data)
		if err != nil {
			return nil, err
		}

		return loggerFromJSON.AttachToContext(ctx), nil
	}

	return ctx, nil
}

func (l *ContextPropogator) InjectFromWorkflow(context workflow.Context, writer workflow.HeaderWriter) error {
	lw, ok := context.Value(ContextKey).(*LogWrapper)

	if ok && lw != nil {
		jsb, err := json.Marshal(lw)
		if err != nil {
			return err
		}

		pl, err := converter.GetDefaultDataConverter().ToPayload(jsb)
		if err != nil {
			return err
		}

		writer.Set(string(ContextKey), pl)
	}

	return nil
}

func (l *ContextPropogator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if pl, ok := reader.Get(string(ContextKey)); ok {
		err := l.logger.UnmarshalJSONSpecial(pl.Data)
		if err != nil {
			return ctx, err
		}

		ctx = workflow.WithValue(ctx, ContextKey, l.logger)
	}

	return ctx, nil
}
