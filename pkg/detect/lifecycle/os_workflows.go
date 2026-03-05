package lifecycle

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// OSLifecycleResult holds the combined result of the OS lifecycle workflow.
type OSLifecycleResult struct {
	MatchResult   MatchOSLifecycleResult
	CleanupResult CleanupStaleOSResult
}

// OSLifecycleWorkflow orchestrates the OS lifecycle detection pipeline.
func OSLifecycleWorkflow(ctx workflow.Context) (*OSLifecycleResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting OSLifecycleWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	runTimestamp := workflow.Now(ctx)

	// 1. Match OS lifecycle.
	var matchResult MatchOSLifecycleResult
	if err := workflow.ExecuteActivity(activityCtx, MatchOSLifecycleActivity,
		MatchOSLifecycleParams{RunTimestamp: runTimestamp}).Get(ctx, &matchResult); err != nil {
		return nil, err
	}
	logger.Info("MatchOSLifecycle done", "matched", matchResult.Matched, "unmatched", matchResult.Unmatched)

	// 2. Cleanup stale rows.
	var cleanupResult CleanupStaleOSResult
	if err := workflow.ExecuteActivity(activityCtx, CleanupStaleOSActivity,
		CleanupStaleOSParams{RunTimestamp: runTimestamp}).Get(ctx, &cleanupResult); err != nil {
		return nil, err
	}
	logger.Info("CleanupStaleOS done", "deleted", cleanupResult.Deleted)

	result := &OSLifecycleResult{
		MatchResult:   matchResult,
		CleanupResult: cleanupResult,
	}

	logger.Info("OSLifecycleWorkflow complete",
		"matched", matchResult.Matched,
		"unmatched", matchResult.Unmatched,
		"deleted", cleanupResult.Deleted)

	return result, nil
}
