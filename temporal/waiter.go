package fcttemporal

import (
	"context"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/maybe"
	"github.com/friendsofgo/errors"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

type FutureError struct {
	Err string `json:"err"`
}

type Future[T any] struct {
	Key       string                   `json:"key"`
	Finalized bool                     `json:"finalized"`
	Data      T                        `json:"data,omitempty"`
	Error     maybe.Maybe[FutureError] `json:"error,omitempty"`
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

func (w *Future[T]) FinalizeWithData(data T) {
	w.Finalized = true
	w.Data = data
}

func (w *Future[T]) FinalizeErr(err error) {
	w.Finalized = true
	w.Error = maybe.WithValue(FutureError{Err: err.Error()})
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
			var queryErr *serviceerror.QueryFailed

			if errors.As(err, &queryErr) && strings.Contains(queryErr.Message, "unknown queryType") {
				// Can be safely ignored, future is not initialized yet
				time.Sleep(250 * time.Millisecond)
				continue
			}
			return err

		}

		var waiter Future[struct{}]

		err = val.Get(&waiter)
		if ctxErr != nil {
			return err
		}

		if waiter.Finalized {
			if futureErr, hasErr := waiter.Error.Value(); hasErr {
				return errors.New(futureErr.Err)
			} else {
				return nil
			}
		}

		time.Sleep(250 * time.Millisecond)
	}
}

// AwaitTypedFuture will poll a workflow for the result of a future, and return it or an error
// the context variable will be used for timeouts
func AwaitTypedFuture[T any](ctx context.Context, temporal client.Client, wfID, runID string, key string) (T, error) {
	var emptyT T

	for {
		ctxErr := ctx.Err()

		if errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return emptyT, ctxErr
		}

		if ctxErr != nil {
			return emptyT, ctxErr
		}

		encodedValue, err := temporal.QueryWorkflow(ctx, wfID, runID, fmt.Sprintf("workflow_waiter_%s", key))
		if err != nil {
			var queryErr *serviceerror.QueryFailed
			if errors.As(err, &queryErr) && strings.Contains(queryErr.Message, "unknown queryType") {
				// Can be safely ignored, future is not initialized yet
				time.Sleep(250 * time.Millisecond)
				continue
			}

			return emptyT, err
		}

		var waiter Future[T]

		err = encodedValue.Get(&waiter)
		if err != nil {
			return emptyT, err
		}

		if waiter.Finalized {
			if futureErr, hasErr := waiter.Error.Value(); hasErr {
				return emptyT, errors.New(futureErr.Err)
			}

			return waiter.Data, nil
		}

		time.Sleep(250 * time.Millisecond)
	}
}
