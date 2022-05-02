package fcttemporal

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"os"
	"strings"
	"testing"
	"time"

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

const StandardHeartbeatSpacing = 10 * time.Second

// ExecuteActivity is a replacement/wrapper for Temporal's built in workflow.ExecuteActivity function, but it allows
// for easier capturing of result values using generics
func ExecuteActivity[T any](ctx workflow.Context, activity any, args ...any) (*T, error) {
	var result T
	err := workflow.ExecuteActivity(ctx, activity, args...).Get(ctx, &result)
	if err != nil {
		return nil, err
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
		return nil, "", "", err
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
	} else if runID == "" {
		panic("empty run id")
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
		return "", "", err
	}

	var unmarshalled workflowRunIdentifier

	err = json.Unmarshal(decoded, &unmarshalled)
	if err != nil {
		return "", "", err
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

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}
