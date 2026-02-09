package group

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1GroupWorkflowResult contains the result of the group workflow.
type S1GroupWorkflowResult struct {
	GroupCount     int
	DurationMillis int64
}

// S1GroupWorkflow ingests SentinelOne groups.
func S1GroupWorkflow(ctx workflow.Context) (*S1GroupWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1GroupWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestS1GroupsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1GroupsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest groups", "error", err)
		return nil, err
	}

	logger.Info("Completed S1GroupWorkflow", "groupCount", result.GroupCount)

	return &S1GroupWorkflowResult{
		GroupCount:     result.GroupCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
