package server

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeComputeServerWorkflowParams contains parameters for the server workflow.
type GreenNodeComputeServerWorkflowParams struct {
	ProjectID string
}

// GreenNodeComputeServerWorkflowResult contains the result of the server workflow.
type GreenNodeComputeServerWorkflowResult struct {
	ServerCount    int
	DurationMillis int64
}

// GreenNodeComputeServerWorkflow ingests GreenNode servers.
func GreenNodeComputeServerWorkflow(ctx workflow.Context, params GreenNodeComputeServerWorkflowParams) (*GreenNodeComputeServerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeServerWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeServersResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeServersActivity, IngestComputeServersParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest servers", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeComputeServerWorkflow",
		"serverCount", result.ServerCount,
	)

	return &GreenNodeComputeServerWorkflowResult{
		ServerCount:    result.ServerCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
