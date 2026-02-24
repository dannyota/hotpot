package volumetype

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeVolumeVolumeTypeWorkflowParams contains parameters for the volume type workflow.
type GreenNodeVolumeVolumeTypeWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeVolumeVolumeTypeWorkflowResult contains the result of the volume type workflow.
type GreenNodeVolumeVolumeTypeWorkflowResult struct {
	VolumeTypeCount int
	DurationMillis  int64
}

// GreenNodeVolumeVolumeTypeWorkflow ingests GreenNode volume types.
func GreenNodeVolumeVolumeTypeWorkflow(ctx workflow.Context, params GreenNodeVolumeVolumeTypeWorkflowParams) (*GreenNodeVolumeVolumeTypeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeVolumeVolumeTypeWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestVolumeVolumeTypesResult
	err := workflow.ExecuteActivity(activityCtx, IngestVolumeVolumeTypesActivity, IngestVolumeVolumeTypesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest volume types", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeVolumeVolumeTypeWorkflow", "volumeTypeCount", result.VolumeTypeCount)

	return &GreenNodeVolumeVolumeTypeWorkflowResult{
		VolumeTypeCount: result.VolumeTypeCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
