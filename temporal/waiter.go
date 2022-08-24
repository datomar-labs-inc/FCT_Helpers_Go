package fcttemporal

import (
	"context"
	"fmt"
	"github.com/friendsofgo/errors"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"time"
)

type Future[T any] struct {
	Key       string `json:"key"`
	Finalized bool   `json:"finalized"`
	Data      *T     `json:"data,omitempty"`
	Error     error  `json:"error,omitempty"`
}

func NewFuture(ctx workflow.Context, key string) *Future[struct{}] {
	w := &Future[struct{}]{
		Key:       fmt.Sprintf("workflow_waiter_%s", key),
		Finalized: false,
	}

	_ = workflow.SetQueryHandler(ctx, w.Key, w.QueryHandler)

	return w
}

func NewTypedFuture[T any](ctx workflow.Context, key string) *Future[T] {
	w := &Future[T]{
		Key:       fmt.Sprintf("workflow_waiter_%s", key),
		Finalized: false,
	}

	_ = workflow.SetQueryHandler(ctx, w.Key, w.QueryHandler)

	return w
}

func (w *Future[T]) QueryHandler() (*Future[T], error) {
	return w, nil
}

func (w *Future[T]) Finalize() {
	w.Finalized = true
}

func (w *Future[T]) FinalizeWithData(data *T) {
	w.Finalized = true
	w.Data = data
}

func (w *Future[T]) FinalizeErr(err error) {
	w.Finalized = true
	w.Error = err
}

// AwaitFuture will poll a workflow for the result of a future, and return it or an error
// the context variable will be used for timeouts
func AwaitFuture(ctx context.Context, temporal client.Client, workflowSingleKey string, key string) error {
	wfID, runID := MustParseWorkflowSingleID(workflowSingleKey)

	for {
		ctxErr := ctx.Err()

		if errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return ctxErr
		}

		val, err := temporal.QueryWorkflow(ctx, wfID, runID, fmt.Sprintf("workflow_waiter_%s", key))
		if err != nil {
			return err
		}

		var waiter Future[struct{}]

		err = val.Get(&waiter)
		if ctxErr != nil {
			return err
		}

		if waiter.Finalized {
			return waiter.Error
		}

		time.Sleep(250 * time.Millisecond)
	}
}

// AwaitTypedFuture will poll a workflow for the result of a future, and return it or an error
// the context variable will be used for timeouts
func AwaitTypedFuture[T any](ctx context.Context, temporal client.Client, workflowSingleKey string, key string) (*T, error) {
	wfID, runID := MustParseWorkflowSingleID(workflowSingleKey)

	for {
		ctxErr := ctx.Err()

		if errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return nil, ctxErr
		}

		val, err := temporal.QueryWorkflow(ctx, wfID, runID, fmt.Sprintf("workflow_waiter_%s", key))
		if err != nil {
			return nil, err
		}

		var waiter Future[T]

		err = val.Get(&waiter)
		if ctxErr != nil {
			return nil, err
		}

		if waiter.Finalized {
			return waiter.Data, waiter.Error
		}

		time.Sleep(250 * time.Millisecond)
	}
}
