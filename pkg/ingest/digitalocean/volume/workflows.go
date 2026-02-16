package volume

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOVolumeWorkflowResult contains the result of the Volume workflow.
type DOVolumeWorkflowResult struct {
	VolumeCount    int
	DurationMillis int64
}

// DOVolumeWorkflow ingests DigitalOcean Volumes.
func DOVolumeWorkflow(ctx workflow.Context) (*DOVolumeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOVolumeWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestDOVolumesResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOVolumesActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest volumes", "error", err)
		return nil, err
	}

	logger.Info("Completed DOVolumeWorkflow", "volumeCount", result.VolumeCount)

	return &DOVolumeWorkflowResult{
		VolumeCount:    result.VolumeCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
