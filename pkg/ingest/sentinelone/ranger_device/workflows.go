package ranger_device

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// S1RangerDeviceWorkflowResult contains the result of the ranger device workflow.
type S1RangerDeviceWorkflowResult struct {
	DeviceCount    int
	DurationMillis int64
}

// S1RangerDeviceWorkflow ingests SentinelOne ranger devices.
func S1RangerDeviceWorkflow(ctx workflow.Context) (*S1RangerDeviceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1RangerDeviceWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestS1RangerDevicesResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1RangerDevicesActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest ranger devices", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed S1RangerDeviceWorkflow", "deviceCount", result.DeviceCount)

	return &S1RangerDeviceWorkflowResult{
		DeviceCount:    result.DeviceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
