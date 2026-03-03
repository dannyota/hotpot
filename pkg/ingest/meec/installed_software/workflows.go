package installed_software

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

const batchSize = 50

// MEECInstalledSoftwareWorkflowResult contains the result of the installed software workflow.
type MEECInstalledSoftwareWorkflowResult struct {
	InstalledSoftwareCount int
	DurationMillis         int64
}

// MEECInstalledSoftwareWorkflow ingests MEEC installed software in batches.
func MEECInstalledSoftwareWorkflow(ctx workflow.Context) (*MEECInstalledSoftwareWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting MEECInstalledSoftwareWorkflow")

	startTime := workflow.Now(ctx)

	// Step 1: List computer IDs from database
	listCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	var listResult ListComputerIDsResult
	if err := workflow.ExecuteActivity(listCtx, ListComputerIDsActivity).Get(ctx, &listResult); err != nil {
		logger.Error("Failed to list computer IDs", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Listed computer IDs", "computerCount", len(listResult.ComputerIDs))

	// Step 2: Process computers in sequential batches.
	// Each batch activity processes up to batchSize computers, with the rate
	// limiter pacing API calls. This avoids fan-out complexity and keeps
	// workflow history small.
	fetchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	})

	totalSoftware := 0
	for i := 0; i < len(listResult.ComputerIDs); i += batchSize {
		end := i + batchSize
		if end > len(listResult.ComputerIDs) {
			end = len(listResult.ComputerIDs)
		}
		batch := listResult.ComputerIDs[i:end]

		var result FetchAndSaveBatchResult
		err := workflow.ExecuteActivity(fetchCtx, FetchAndSaveBatchActivity, FetchAndSaveBatchInput{
			ComputerIDs: batch,
			CollectedAt: listResult.CollectedAt,
		}).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to process batch", "batchStart", i, "error", err)
			return nil, temporalerr.PropagateNonRetryable(err)
		}

		totalSoftware += result.SoftwareCount
		logger.Info("meec installed software: batch complete",
			"computersDone", end,
			"totalComputers", len(listResult.ComputerIDs),
			"totalSoftware", totalSoftware,
		)
	}

	// Step 3: Delete orphan installed software
	deleteCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	if err := workflow.ExecuteActivity(deleteCtx, DeleteOrphanInstalledSoftwareActivity).Get(ctx, nil); err != nil {
		logger.Warn("Failed to delete orphan installed software", "error", err)
	}

	durationMillis := workflow.Now(ctx).Sub(startTime).Milliseconds()
	logger.Info("Completed MEECInstalledSoftwareWorkflow", "installedSoftwareCount", totalSoftware, "durationMillis", durationMillis)

	return &MEECInstalledSoftwareWorkflowResult{
		InstalledSoftwareCount: totalSoftware,
		DurationMillis:         durationMillis,
	}, nil
}
