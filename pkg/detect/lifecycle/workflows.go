package lifecycle

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// SoftwareLifecycleResult holds the combined result of the workflow.
type SoftwareLifecycleResult struct {
	MatchResult     MatchProductsResult
	OSCoreResult    ClassifyOSCoreResult
	UnmatchedResult MarkUnmatchedResult
	CleanupResult   CleanupStaleResult
}

// SoftwareLifecycleWorkflow orchestrates the 4-activity lifecycle detection pipeline.
func SoftwareLifecycleWorkflow(ctx workflow.Context) (*SoftwareLifecycleResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting SoftwareLifecycleWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
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

	// 1. Match products.
	var matchResult MatchProductsResult
	if err := workflow.ExecuteActivity(activityCtx, MatchProductsActivity,
		MatchProductsParams{RunTimestamp: runTimestamp}).Get(ctx, &matchResult); err != nil {
		return nil, err
	}
	logger.Info("MatchProducts done", "matched", matchResult.Matched, "names", len(matchResult.MatchedNames))

	// 2. Classify OS core.
	var osCoreResult ClassifyOSCoreResult
	if err := workflow.ExecuteActivity(activityCtx, ClassifyOSCoreActivity,
		ClassifyOSCoreParams{
			RunTimestamp: runTimestamp,
			MatchedNames: matchResult.MatchedNames,
		}).Get(ctx, &osCoreResult); err != nil {
		return nil, err
	}
	logger.Info("ClassifyOSCore done", "os_core", osCoreResult.OSCore, "names", len(osCoreResult.OSCoreNames))

	// 3. Mark unmatched.
	var unmatchedResult MarkUnmatchedResult
	if err := workflow.ExecuteActivity(activityCtx, MarkUnmatchedActivity,
		MarkUnmatchedParams{
			RunTimestamp: runTimestamp,
			MatchedNames: matchResult.MatchedNames,
			OSCoreNames:  osCoreResult.OSCoreNames,
		}).Get(ctx, &unmatchedResult); err != nil {
		return nil, err
	}
	logger.Info("MarkUnmatched done", "unmatched", unmatchedResult.Unmatched)

	// 4. Cleanup stale rows.
	var cleanupResult CleanupStaleResult
	if err := workflow.ExecuteActivity(activityCtx, CleanupStaleActivity,
		CleanupStaleParams{RunTimestamp: runTimestamp}).Get(ctx, &cleanupResult); err != nil {
		return nil, err
	}
	logger.Info("CleanupStale done", "deleted", cleanupResult.Deleted)

	result := &SoftwareLifecycleResult{
		MatchResult:     matchResult,
		OSCoreResult:    osCoreResult,
		UnmatchedResult: unmatchedResult,
		CleanupResult:   cleanupResult,
	}

	logger.Info("SoftwareLifecycleWorkflow complete",
		"matched", matchResult.Matched,
		"os_core", osCoreResult.OSCore,
		"unmatched", unmatchedResult.Unmatched,
		"deleted", cleanupResult.Deleted)

	return result, nil
}
