package osimage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeComputeOSImageWorkflowParams contains parameters for the OS image workflow.
type GreenNodeComputeOSImageWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeComputeOSImageWorkflowResult contains the result of the OS image workflow.
type GreenNodeComputeOSImageWorkflowResult struct {
	OSImageCount   int
	DurationMillis int64
}

// GreenNodeComputeOSImageWorkflow ingests GreenNode OS images.
func GreenNodeComputeOSImageWorkflow(ctx workflow.Context, params GreenNodeComputeOSImageWorkflowParams) (*GreenNodeComputeOSImageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeOSImageWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestComputeOSImagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeOSImagesActivity, IngestComputeOSImagesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest OS images", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeComputeOSImageWorkflow", "osImageCount", result.OSImageCount)

	return &GreenNodeComputeOSImageWorkflowResult{
		OSImageCount:   result.OSImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
