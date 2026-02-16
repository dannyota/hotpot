package logexclusion

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPLoggingLogExclusionWorkflowParams contains parameters for the log exclusion workflow.
type GCPLoggingLogExclusionWorkflowParams struct {
	ProjectID string
}

// GCPLoggingLogExclusionWorkflowResult contains the result of the log exclusion workflow.
type GCPLoggingLogExclusionWorkflowResult struct {
	ProjectID      string
	ExclusionCount int
	DurationMillis int64
}

// GCPLoggingLogExclusionWorkflow ingests GCP Cloud Logging log exclusions for a single project.
func GCPLoggingLogExclusionWorkflow(ctx workflow.Context, params GCPLoggingLogExclusionWorkflowParams) (*GCPLoggingLogExclusionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPLoggingLogExclusionWorkflow", "projectID", params.ProjectID)

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

	var result IngestLoggingLogExclusionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoggingLogExclusionsActivity, IngestLoggingLogExclusionsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest log exclusions", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPLoggingLogExclusionWorkflow",
		"projectID", params.ProjectID,
		"exclusionCount", result.ExclusionCount,
	)

	return &GCPLoggingLogExclusionWorkflowResult{
		ProjectID:      result.ProjectID,
		ExclusionCount: result.ExclusionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
