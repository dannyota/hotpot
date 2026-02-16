package accesslevel

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAccessContextManagerAccessLevelWorkflowParams contains parameters for the access level workflow.
type GCPAccessContextManagerAccessLevelWorkflowParams struct {
}

// GCPAccessContextManagerAccessLevelWorkflowResult contains the result of the access level workflow.
type GCPAccessContextManagerAccessLevelWorkflowResult struct {
	LevelCount int
}

// GCPAccessContextManagerAccessLevelWorkflow ingests access levels.
func GCPAccessContextManagerAccessLevelWorkflow(ctx workflow.Context, params GCPAccessContextManagerAccessLevelWorkflowParams) (*GCPAccessContextManagerAccessLevelWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAccessContextManagerAccessLevelWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestAccessLevelsResult
	err := workflow.ExecuteActivity(activityCtx, IngestAccessLevelsActivity, IngestAccessLevelsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest access levels", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAccessContextManagerAccessLevelWorkflow",
		"levelCount", result.LevelCount,
	)

	return &GCPAccessContextManagerAccessLevelWorkflowResult{
		LevelCount: result.LevelCount,
	}, nil
}
