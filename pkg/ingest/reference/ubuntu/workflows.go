package ubuntu

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// UbuntuPackagesWorkflowResult contains the result of the Ubuntu packages workflow.
type UbuntuPackagesWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// UbuntuPackagesWorkflow ingests Ubuntu package indexes.
func UbuntuPackagesWorkflow(ctx workflow.Context) (*UbuntuPackagesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting UbuntuPackagesWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestUbuntuPackagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestUbuntuPackagesActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Ubuntu packages", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed UbuntuPackagesWorkflow", "packageCount", result.PackageCount)

	return &UbuntuPackagesWorkflowResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
