package userimage

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeComputeUserImageWorkflowParams contains parameters for the user image workflow.
type GreenNodeComputeUserImageWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeComputeUserImageWorkflowResult contains the result of the user image workflow.
type GreenNodeComputeUserImageWorkflowResult struct {
	UserImageCount int
	DurationMillis int64
}

// GreenNodeComputeUserImageWorkflow ingests GreenNode user images.
func GreenNodeComputeUserImageWorkflow(ctx workflow.Context, params GreenNodeComputeUserImageWorkflowParams) (*GreenNodeComputeUserImageWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeUserImageWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestComputeUserImagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeUserImagesActivity, IngestComputeUserImagesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest user images", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeComputeUserImageWorkflow", "userImageCount", result.UserImageCount)

	return &GreenNodeComputeUserImageWorkflowResult{
		UserImageCount: result.UserImageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
