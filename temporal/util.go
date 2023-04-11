package fcttemporal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datomar-labs-inc/FCT_Helpers_Go/ferr"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"go.temporal.io/sdk/workflow"
)

const (
	ActivityTimeoutXS = 8 * time.Second
	ActivityTimeoutS  = 25 * time.Second
	ActivityTimeoutM  = 1 * time.Minute
	ActivityTimeoutL  = 10 * time.Minute
	ActivityTimeoutXL = 25 * time.Minute
)

var (
	ActivityMaxRetriesNone = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    1,
	}

	// ActivityMaxRetriesS 3 retries
	ActivityMaxRetriesS = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    ActivityTimeoutM,
		MaximumAttempts:    3,
	}

	// ActivityMaxRetriesM 15 retries
	ActivityMaxRetriesM = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    ActivityTimeoutXL,
		MaximumAttempts:    15,
	}

	ActivityMaxRetriesL = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    ActivityTimeoutXL,
		MaximumAttempts:    75,
	}

	ActivityMaxRetriesUnlimited = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    ActivityTimeoutXL,
	}
)

var (
	ActivityCtxSmall = func(ctx workflow.Context) workflow.Context {
		return ActivityCtx(ctx, ActivityTimeoutS, ActivityMaxRetriesS)
	}
	LocalActivityCtxSmall = func(ctx workflow.Context) workflow.Context {
		return LocalActivityCtx(ctx, ActivityTimeoutS, ActivityMaxRetriesS)
	}
	LocalActivityCtxExtraSmall = func(ctx workflow.Context) workflow.Context {
		return LocalActivityCtx(ctx, ActivityTimeoutXS, ActivityMaxRetriesNone)
	}
)

const StandardHeartbeatSpacing = 10 * time.Second

func ActivitySmall[T any](ctx workflow.Context, activity any, args ...any) (T, error) {
	var result T

	err := workflow.ExecuteActivity(ActivityCtxSmall(ctx), activity, args...).Get(ctx, &result)
	if err != nil {
		return result, ferr.Wrap(err)
	}

	return result, nil
}

// ExecuteActivity is a replacement/wrapper for Temporal's built in workflow.ExecuteActivity function, but it allows
// for easier capturing of result values using generics
func ExecuteActivity[T any](ctx workflow.Context, activity any, args ...any) (*T, error) {
	var result T
	err := workflow.ExecuteActivity(ctx, activity, args...).Get(ctx, &result)
	if err != nil {
		return nil, ferr.Wrap(err)
	}

	return &result, nil
}

// ActivityCtx is a helper to create a context with workflow.ActivityOptions attached
// The activity options will be created with a max run duration of timeout
func ActivityCtx(ctx workflow.Context, timeout time.Duration, retry *temporal.RetryPolicy) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		RetryPolicy:         retry,
	})
} // LocalActivityCtx is a helper to create a context with workflow.LocalActivityOptions attached
// The activity options will be created with a max run duration of timeout
func LocalActivityCtx(ctx workflow.Context, timeout time.Duration, retry *temporal.RetryPolicy) workflow.Context {
	return workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		StartToCloseTimeout: timeout,
		RetryPolicy:         retry,
	})
}

// ActivityCtxWithTaskQueue is a helper to create a context with workflow.ActivityOptions attached
// The activity options will be created with a max run duration of timeout
// The activity will be run in the given TaskQueue (for inter worker workflows)
func ActivityCtxWithTaskQueue(ctx workflow.Context, timeout time.Duration, retry *temporal.RetryPolicy, taskQueue string) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		RetryPolicy:         retry,
		TaskQueue:           taskQueue,
	})
}

// ActivityCtxDC is a helper to create a disconnected context with workflow.ActivityOptions attached
// The activity options will be created with a max run duration of timeout
func ActivityCtxDC(ctx workflow.Context, timeout time.Duration) workflow.Context {
	dCtx, _ := workflow.NewDisconnectedContext(ctx)

	return workflow.WithActivityOptions(dCtx, workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
	})
}

// ActivityCtxHB is a helper to create a context with workflow.ActivityOptions attached
// The activity options will be created with a max run duration of timeout
// And a default heartbeat spacing of StandardHeartbeatSpacing
func ActivityCtxHB(ctx workflow.Context, timeout time.Duration, retry *temporal.RetryPolicy) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: timeout,
		HeartbeatTimeout:    StandardHeartbeatSpacing,
		RetryPolicy:         retry,
	})
}

func ExecuteWorkflowSync[T any](ctx context.Context, temporalClient client.Client, options client.StartWorkflowOptions, workflow any, args ...any) (result *T, workfowID string, runID string, err error) {
	wfRun, err := temporalClient.ExecuteWorkflow(ctx, options, workflow, args...)
	if err != nil {
		return nil, "", "", ferr.Wrap(err)
	}

	workfowID = wfRun.GetID()
	runID = wfRun.GetRunID()

	err = wfRun.Get(ctx, result)
	if err != nil {
		return
	}

	return
}

type workflowRunIdentifier struct {
	WorkflowID string `json:"wid"`
	RunID      string `json:"rid"`
}

func MustGetWorkflowSingleID(workflowID, runID string) string {
	workflowID = strings.TrimSpace(workflowID)
	runID = strings.TrimSpace(runID)

	if workflowID == "" {
		panic("empty workflow id")
	}

	marshalled, err := json.Marshal(workflowRunIdentifier{
		WorkflowID: workflowID,
		RunID:      runID,
	})
	if err != nil {
		panic(err)
	}

	return base64.RawURLEncoding.EncodeToString(marshalled)
}

func MustExtractWorkflowSingleID(ctx workflow.Context) string {
	info := workflow.GetInfo(ctx)
	return MustGetWorkflowSingleID(info.WorkflowExecution.ID, info.WorkflowExecution.RunID)
}

func ParseWorkflowSingleID(id string) (workflowID string, runID string, err error) {

	decoded, err := base64.RawURLEncoding.DecodeString(id)
	if err != nil {
		return "", "", ferr.Wrap(err)
	}

	var unmarshalled workflowRunIdentifier

	err = json.Unmarshal(decoded, &unmarshalled)
	if err != nil {
		return "", "", ferr.Wrap(err)
	}

	return unmarshalled.WorkflowID, unmarshalled.RunID, nil
}

func MustParseWorkflowSingleID(id string) (workflowID string, runID string) {
	wfid, rid, err := ParseWorkflowSingleID(id)
	if err != nil {
		panic(err)
	}

	return wfid, rid
}

func Receive[T any](ctx workflow.Context, ch workflow.ReceiveChannel) *T {
	var result T

	// TODO handle more

	_ = ch.Receive(ctx, &result)

	return &result
}

type SignalSwitch struct {
	Selector       workflow.Selector
	SignalFired    string
	errChan        chan error
	handlerDidFire bool

	wrappedReceiverRegistered bool
	wrappedHandlerFuncs       map[string]func(workflow.Context, []byte)
}

func NewSignalSwitch(ctx workflow.Context) *SignalSwitch {
	return &SignalSwitch{
		Selector:            workflow.NewSelector(ctx),
		errChan:             make(chan error, 1),
		wrappedHandlerFuncs: map[string]func(workflow.Context, []byte){},
	}
}

type SignalWrapper[T any] struct {
	SignalType string `json:"signal_type"`
	Signal     T      `json:"signal"`
}

func (ss *SignalSwitch) Select(ctx workflow.Context) error {
	ss.Selector.Select(ctx)

	if !ss.handlerDidFire {
		return nil
	}

	return <-ss.errChan
}

const SignalSwitchSignalType = "signal_switch_signal"

func WrapSignal[T any](signalType string, signal T) SignalWrapper[T] {
	return SignalWrapper[T]{
		SignalType: signalType,
		Signal:     signal,
	}
}

func AddSignalHandler[T any](ctx workflow.Context, ss *SignalSwitch, signal string, handler SignalHandler[T]) {
	if !ss.wrappedReceiverRegistered {
		ss.Selector.AddReceive(workflow.GetSignalChannel(ctx, SignalSwitchSignalType), func(c workflow.ReceiveChannel, more bool) {
			var temp any
			var wrappedSignalType string

			c.Receive(ctx, &temp)

			switch temp.(type) {
			case map[string]any:
				if signalType, ok := temp.(map[string]any)["signal_type"].(string); ok {
					wrappedSignalType = signalType
				} else {
					ss.errChan <- fmt.Errorf("signal type is not a string")
					return
				}
			default:
				ss.errChan <- fmt.Errorf("unexpected signal type %T", temp)
				return
			}

			if handlerFunc, ok := ss.wrappedHandlerFuncs[wrappedSignalType]; ok {
				signalValueJSON, err := json.Marshal(temp)
				if err != nil {
					ss.errChan <- err
					return
				}

				handlerFunc(ctx, signalValueJSON)
			} else {
				ss.errChan <- fmt.Errorf("no handler for signal %s", ss.SignalFired)
			}
		})

		ss.wrappedReceiverRegistered = true
	}

	ss.wrappedHandlerFuncs[signal] = func(ctx workflow.Context, signalValueJSON []byte) {
		var signalValue SignalWrapper[T]

		err := json.Unmarshal(signalValueJSON, &signalValue)
		if err != nil {
			ss.errChan <- err
			return
		}

		ss.SignalFired = signal
		ss.handlerDidFire = true

		err = handler(ctx, signalValue.Signal)
		if err != nil {
			ss.errChan <- err
		} else {
			ss.errChan <- nil
		}
	}
}

func AddUnwrappedSignalHandler[T any](ctx workflow.Context, ss *SignalSwitch, signal string, handler SignalHandler[T]) {
	ss.Selector.AddReceive(workflow.GetSignalChannel(ctx, signal), func(c workflow.ReceiveChannel, more bool) {
		var signalValue T

		c.Receive(ctx, &signalValue)

		ss.SignalFired = signal
		ss.handlerDidFire = true

		err := handler(ctx, signalValue)
		if err != nil {
			ss.errChan <- err
		} else {
			ss.errChan <- nil
		}
	})
}

func AddFutureHandler[T any](ctx workflow.Context, ss *SignalSwitch, future workflow.Future, handler SignalHandler[T]) {
	ss.Selector.AddFuture(future, func(f workflow.Future) {
		var futureValue T

		err := f.Get(ctx, &futureValue)
		if err != nil {
			ss.errChan <- err
			return
		}

		ss.handlerDidFire = true

		err = handler(ctx, futureValue)
		if err != nil {
			ss.errChan <- err
		} else {
			ss.errChan <- nil
		}
	})
}

type SignalHandler[T any] func(ctx workflow.Context, signalBody T) error
