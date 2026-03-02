package software

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// MEECSoftwareWorkflowResult contains the result of the software workflow.
type MEECSoftwareWorkflowResult struct {
	SoftwareCount  int
	DurationMillis int64
}

// MEECSoftwareWorkflow ingests the MEEC software catalog.
func MEECSoftwareWorkflow(ctx workflow.Context) (*MEECSoftwareWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting MEECSoftwareWorkflow")

	startTime := workflow.Now(ctx)

	// Step 1: Ingest software catalog
	ingestCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	var ingestResult IngestSoftwareResult
	err := workflow.ExecuteActivity(ingestCtx, IngestSoftwareActivity).Get(ctx, &ingestResult)
	if err != nil {
		logger.Error("Failed to ingest software", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Ingested software catalog", "softwareCount", ingestResult.SoftwareCount)

	// Step 2: Delete stale software entries
	deleteCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	if err := workflow.ExecuteActivity(deleteCtx, DeleteStaleSoftwareActivity, DeleteStaleSoftwareInput{
		CollectedAt: ingestResult.CollectedAt,
	}).Get(ctx, nil); err != nil {
		logger.Warn("Failed to delete stale software entries", "error", err)
	}

	durationMillis := workflow.Now(ctx).Sub(startTime).Milliseconds()
	logger.Info("Completed MEECSoftwareWorkflow", "softwareCount", ingestResult.SoftwareCount, "durationMillis", durationMillis)

	return &MEECSoftwareWorkflowResult{
		SoftwareCount:  ingestResult.SoftwareCount,
		DurationMillis: durationMillis,
	}, nil
}
