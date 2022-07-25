package lggr

import (
	"context"
	"encoding/json"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

func NewContextPropagator() *lggrContextPropagator {
	return &lggrContextPropagator{}
}

type lggrContextPropagator struct {}

func (l *lggrContextPropagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
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

		writer.Set(ContextKey, pl)
	}

	return nil
}

func (l *lggrContextPropagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if pl, ok := reader.Get(ContextKey); ok {
		var lggr LogWrapper

		err := json.Unmarshal(pl.Data, &lggr)
		if err != nil {
			return ctx, err
		}

		return lggr.AttachToContext(ctx), nil
	}

	return ctx, nil
}

func (l *lggrContextPropagator) InjectFromWorkflow(context workflow.Context, writer workflow.HeaderWriter) error {
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

		writer.Set(ContextKey, pl)
	}

	return nil
}

func (l *lggrContextPropagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if pl, ok := reader.Get(ContextKey); ok {
		var lggr LogWrapper

		err := json.Unmarshal(pl.Data, &lggr)
		if err != nil {
			return ctx, err
		}

		ctx = workflow.WithValue(ctx, ContextKey, &lggr)
	}

	return ctx, nil
}



