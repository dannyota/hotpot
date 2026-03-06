package app_inventory

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// S1AppInventoryWorkflowResult contains the result of the app inventory workflow.
type S1AppInventoryWorkflowResult struct {
	AppCount       int
	DurationMillis int64
}

// S1AppInventoryWorkflow ingests SentinelOne application inventory.
func S1AppInventoryWorkflow(ctx workflow.Context) (*S1AppInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1AppInventoryWorkflow")

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

	var result IngestS1AppInventoryResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1AppInventoryActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest app inventory", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed S1AppInventoryWorkflow", "appCount", result.AppCount)

	return &S1AppInventoryWorkflowResult{
		AppCount:       result.AppCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
