package image

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeImageWorkflowParams contains parameters for the image workflow.
type GCPComputeImageWorkflowParams struct {
	ProjectID string
}

// GCPComputeImageWorkflowResult contains the result of the image workflow.
type GCPComputeImageWorkflowResult struct {
	ProjectID      string
	ImageCount     int
	DurationMillis int64
}

// GCPComputeImageWorkflow ingests GCP Compute images for a single project.
func GCPComputeImageWorkflow(ctx workflow.Context, params GCPComputeImageWorkflowParams) (*GCPComputeImageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeImageWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestComputeImagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeImagesActivity, IngestComputeImagesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest images", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeImageWorkflow",
		"projectID", params.ProjectID,
		"imageCount", result.ImageCount,
	)

	return &GCPComputeImageWorkflowResult{
		ProjectID:      result.ProjectID,
		ImageCount:     result.ImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
