package rpm

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// RPMPackagesWorkflowResult contains the result of the RPM packages workflow.
type RPMPackagesWorkflowResult struct {
	PackageCount   int
	DurationMillis int64
}

// RPMPackagesWorkflow ingests RPM repository metadata.
func RPMPackagesWorkflow(ctx workflow.Context) (*RPMPackagesWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting RPMPackagesWorkflow")

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

	var result IngestRPMPackagesResult
	err := workflow.ExecuteActivity(activityCtx, IngestRPMPackagesActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest RPM packages", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed RPMPackagesWorkflow", "packageCount", result.PackageCount)

	return &RPMPackagesWorkflowResult{
		PackageCount:   result.PackageCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
