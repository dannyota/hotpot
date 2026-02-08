package urlmap

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeUrlMapWorkflowParams contains parameters for the URL map workflow.
type GCPComputeUrlMapWorkflowParams struct {
	ProjectID string
}

// GCPComputeUrlMapWorkflowResult contains the result of the URL map workflow.
type GCPComputeUrlMapWorkflowResult struct {
	ProjectID      string
	UrlMapCount    int
	DurationMillis int64
}

// GCPComputeUrlMapWorkflow ingests GCP Compute URL maps for a single project.
func GCPComputeUrlMapWorkflow(ctx workflow.Context, params GCPComputeUrlMapWorkflowParams) (*GCPComputeUrlMapWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeUrlMapWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeUrlMapsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeUrlMapsActivity, IngestComputeUrlMapsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest URL maps", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeUrlMapWorkflow",
		"projectID", params.ProjectID,
		"urlMapCount", result.UrlMapCount,
	)

	return &GCPComputeUrlMapWorkflowResult{
		ProjectID:      result.ProjectID,
		UrlMapCount:    result.UrlMapCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
