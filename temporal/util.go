package fct_temporal

import (
	"go.temporal.io/sdk/temporal"
	"os"
	"testing"
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
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

	ActivityMaxRetriesS = &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2,
		MaximumInterval:    ActivityTimeoutM,
		MaximumAttempts:    3,
	}

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

func skipCI(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}
}
