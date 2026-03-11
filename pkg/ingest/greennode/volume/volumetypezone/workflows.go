package volumetypezone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeVolumeVolumeTypeZoneWorkflowParams contains parameters for the volume type zone workflow.
type GreenNodeVolumeVolumeTypeZoneWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeVolumeVolumeTypeZoneWorkflowResult contains the result of the volume type zone workflow.
type GreenNodeVolumeVolumeTypeZoneWorkflowResult struct {
	VolumeTypeZoneCount int
	DurationMillis      int64
}

// GreenNodeVolumeVolumeTypeZoneWorkflow ingests GreenNode volume type zones.
func GreenNodeVolumeVolumeTypeZoneWorkflow(ctx workflow.Context, params GreenNodeVolumeVolumeTypeZoneWorkflowParams) (*GreenNodeVolumeVolumeTypeZoneWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeVolumeVolumeTypeZoneWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestVolumeVolumeTypeZonesResult
	err := workflow.ExecuteActivity(activityCtx, IngestVolumeVolumeTypeZonesActivity, IngestVolumeVolumeTypeZonesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest volume type zones", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeVolumeVolumeTypeZoneWorkflow", "volumeTypeZoneCount", result.VolumeTypeZoneCount)

	return &GreenNodeVolumeVolumeTypeZoneWorkflowResult{
		VolumeTypeZoneCount: result.VolumeTypeZoneCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
