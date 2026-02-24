package blockvolume

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeVolumeBlockVolumeWorkflowParams contains parameters for the block volume workflow.
type GreenNodeVolumeBlockVolumeWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeVolumeBlockVolumeWorkflowResult contains the result of the block volume workflow.
type GreenNodeVolumeBlockVolumeWorkflowResult struct {
	BlockVolumeCount int
	DurationMillis   int64
}

// GreenNodeVolumeBlockVolumeWorkflow ingests GreenNode block volumes.
func GreenNodeVolumeBlockVolumeWorkflow(ctx workflow.Context, params GreenNodeVolumeBlockVolumeWorkflowParams) (*GreenNodeVolumeBlockVolumeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeVolumeBlockVolumeWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestVolumeBlockVolumesResult
	err := workflow.ExecuteActivity(activityCtx, IngestVolumeBlockVolumesActivity, IngestVolumeBlockVolumesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest block volumes", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeVolumeBlockVolumeWorkflow",
		"blockVolumeCount", result.BlockVolumeCount,
	)

	return &GreenNodeVolumeBlockVolumeWorkflowResult{
		BlockVolumeCount: result.BlockVolumeCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
