package fcttemporal

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/datomar-labs-inc/FCT_Helpers_Go/maybe"
	"github.com/friendsofgo/errors"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"strings"
	"time"
)

const (
	PubSubVersionChangeKey    = "pubsub_version_change"
	PubSubVersionMaxSupported = 1
)

type FutureError struct {
	IsFutureError bool   `json:"_is_future_error"`
	Err           string `json:"err"`
}

type FutureResolver interface {
	Resolve(ctx context.Context, wfID string, future Future[any]) error
	AwaitContextTimeout(ctx context.Context, key string, timeout time.Duration) ([]byte, error)
	GetTemporalClient() client.Client
}

type Future[T any] struct {
	Key       string                   `json:"key"`
	Finalized bool                     `json:"finalized"`
	Data      T                        `json:"data,omitempty"`
	Error     maybe.Maybe[FutureError] `json:"error,omitempty"`
}

func NewFuture(ctx workflow.Context, key string) *Future[struct{}] {
	w := &Future[struct{}]{
		Key:       key,
		Finalized: false,
	}

	_ = workflow.SetQueryHandler(ctx, fmt.Sprintf("workflow_waiter_%s", w.Key), w.QueryHandler)

	return w
}

func NewTypedFuture[T any](ctx workflow.Context, key string) *Future[T] {
	w := &Future[T]{
		Key:       key,
		Finalized: false,
	}

	_ = workflow.SetQueryHandler(ctx, fmt.Sprintf("workflow_waiter_%s", w.Key), w.QueryHandler)

	return w
}

func (w *Future[T]) QueryHandler() (*Future[T], error) {
	return w, nil
}

func (w *Future[T]) Finalize(ctx workflow.Context, resolver FutureResolver) {
	w.Finalized = true

	v := workflow.GetVersion(ctx, PubSubVersionChangeKey, workflow.DefaultVersion, PubSubVersionMaxSupported)

	if v != workflow.DefaultVersion {
		aCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
			StartToCloseTimeout: ActivityTimeoutXS,
		})

		err := workflow.ExecuteLocalActivity(aCtx, func(cctx context.Context) error {
			err := resolver.Resolve(cctx, workflow.GetInfo(ctx).WorkflowExecution.ID, Future[any]{
				Key:       w.Key,
				Finalized: w.Finalized,
				Data:      w.Data,
				Error:     w.Error,
			})
			if err != nil {
				return err
			}

			return nil
		}).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("error resolving future", "err", err)
		}
	}
}

func (w *Future[T]) FinalizeWithData(ctx workflow.Context, resolver FutureResolver, data T) {
	w.Finalized = true
	w.Data = data

	v := workflow.GetVersion(ctx, PubSubVersionChangeKey, workflow.DefaultVersion, PubSubVersionMaxSupported)

	if v != workflow.DefaultVersion {
		aCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
			StartToCloseTimeout: ActivityTimeoutXS,
		})

		err := workflow.ExecuteLocalActivity(aCtx, func(cctx context.Context) error {
			err := resolver.Resolve(cctx, workflow.GetInfo(ctx).WorkflowExecution.ID, Future[any]{
				Key:       w.Key,
				Finalized: w.Finalized,
				Data:      w.Data,
				Error:     w.Error,
			})
			if err != nil {
				return err
			}

			return nil
		}).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("error resolving future", "err", err)
		}
	}
}

func (w *Future[T]) FinalizeErr(ctx workflow.Context, resolver FutureResolver, err error) {
	w.Finalized = true
	w.Error = maybe.WithValue(FutureError{Err: err.Error()})

	v := workflow.GetVersion(ctx, PubSubVersionChangeKey, workflow.DefaultVersion, PubSubVersionMaxSupported)

	if v != workflow.DefaultVersion {
		aCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
			StartToCloseTimeout: ActivityTimeoutXS,
		})

		err := workflow.ExecuteLocalActivity(aCtx, func(cctx context.Context) error {
			err := resolver.Resolve(cctx, workflow.GetInfo(ctx).WorkflowExecution.ID, Future[any]{
				Key:       w.Key,
				Finalized: w.Finalized,
				Data:      w.Data,
				Error:     w.Error,
			})
			if err != nil {
				return err
			}

			return nil
		}).Get(ctx, nil)
		if err != nil {
			workflow.GetLogger(ctx).Error("error resolving future", "err", err)
		}
	}
}

// AwaitFuture will poll a workflow for the result of a future, and return it or an error
// the context variable will be used for timeouts
func AwaitFuture(ctx context.Context, resolver FutureResolver, workflowSingleKey string, key string) error {
	wfID, runID := MustParseWorkflowSingleID(workflowSingleKey)

	dataBytes, err := resolver.AwaitContextTimeout(ctx, fmt.Sprintf("%s:%s", wfID, key), 5*time.Second)
	if err == nil {
		var output Future[struct{}]

		err = json.Unmarshal(dataBytes, &output)
		if err != nil {
			return err
		}

		if output.Error.HasValue() {
			return errors.New(output.Error.Or(FutureError{Err: "unknown error"}).Err)
		}

		return nil
	}

	for {
		ctxErr := ctx.Err()

		if errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return ctxErr
		}

		val, err := resolver.GetTemporalClient().QueryWorkflow(ctx, wfID, runID, fmt.Sprintf("workflow_waiter_%s", key))
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
			}

			return nil
		}

		time.Sleep(250 * time.Millisecond)
	}
}

// AwaitTypedFuture will poll a workflow for the result of a future, and return it or an error
// the context variable will be used for timeouts
func AwaitTypedFuture[T any](ctx context.Context, resolver FutureResolver, wfID, runID string, key string) (T, error) {
	var emptyT T

	dataBytes, err := resolver.AwaitContextTimeout(ctx, fmt.Sprintf("%s:%s", wfID, key), 5*time.Second)
	if err == nil {
		var output Future[T]

		err = json.Unmarshal(dataBytes, &output)
		if err != nil {
			return emptyT, err
		}

		if output.Error.HasValue() {
			return emptyT, errors.New(output.Error.Or(FutureError{Err: "unknown error"}).Err)
		}

		return output.Data, nil
	}

	for {
		ctxErr := ctx.Err()

		if errors.Is(ctxErr, context.Canceled) || errors.Is(ctxErr, context.DeadlineExceeded) {
			return emptyT, ctxErr
		}

		if ctxErr != nil {
			return emptyT, ctxErr
		}

		encodedValue, err := resolver.GetTemporalClient().QueryWorkflow(ctx, wfID, runID, fmt.Sprintf("workflow_waiter_%s", key))
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
