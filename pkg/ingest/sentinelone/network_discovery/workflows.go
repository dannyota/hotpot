package network_discovery

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// S1NetworkDiscoveryWorkflowResult contains the result of the network discovery workflow.
type S1NetworkDiscoveryWorkflowResult struct {
	DeviceCount    int
	DurationMillis int64
}

// S1NetworkDiscoveryWorkflow ingests SentinelOne XDR network discovery devices.
func S1NetworkDiscoveryWorkflow(ctx workflow.Context) (*S1NetworkDiscoveryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1NetworkDiscoveryWorkflow")

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

	var result IngestS1NetworkDiscoveryResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1NetworkDiscoveryActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest network discovery devices", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed S1NetworkDiscoveryWorkflow", "deviceCount", result.DeviceCount)

	return &S1NetworkDiscoveryWorkflowResult{
		DeviceCount:    result.DeviceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
