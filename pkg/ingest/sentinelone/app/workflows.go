package app

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1AppWorkflowResult contains the result of the app workflow.
type S1AppWorkflowResult struct {
	AppCount       int
	DurationMillis int64
}

// S1AppWorkflow ingests SentinelOne installed applications.
func S1AppWorkflow(ctx workflow.Context) (*S1AppWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1AppWorkflow")

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

	var result IngestS1AppsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1AppsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest apps", "error", err)
		return nil, err
	}

	logger.Info("Completed S1AppWorkflow", "appCount", result.AppCount)

	return &S1AppWorkflowResult{
		AppCount:       result.AppCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
