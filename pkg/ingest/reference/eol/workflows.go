package eol

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// EOLWorkflowResult contains the result of the EOL workflow.
type EOLWorkflowResult struct {
	ProductCount   int
	CycleCount     int
	DurationMillis int64
}

// EOLWorkflow ingests the endoflife.date database.
func EOLWorkflow(ctx workflow.Context) (*EOLWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting EOLWorkflow")

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

	// Step 1: Ingest endoflife.date (full replace).
	var eolResult IngestEOLResult
	err := workflow.ExecuteActivity(activityCtx, IngestEOLActivity).Get(ctx, &eolResult)
	if err != nil {
		logger.Error("Failed to ingest EOL data", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	// Step 2: Ingest RHEL EUS cycles from Red Hat (inserts after full replace).
	var eusResult IngestRHELEUSResult
	err = workflow.ExecuteActivity(activityCtx, IngestRHELEUSActivity).Get(ctx, &eusResult)
	if err != nil {
		logger.Error("Failed to ingest RHEL EUS data", "error", err)
		// Non-fatal: continue with main EOL data.
	} else {
		logger.Info("Ingested RHEL EUS cycles", "cycleCount", eusResult.CycleCount)
	}

	totalCycles := eolResult.CycleCount + eusResult.CycleCount

	logger.Info("Completed EOLWorkflow",
		"productCount", eolResult.ProductCount,
		"cycleCount", totalCycles,
	)

	return &EOLWorkflowResult{
		ProductCount:   eolResult.ProductCount,
		CycleCount:     totalCycles,
		DurationMillis: eolResult.DurationMillis + eusResult.DurationMillis,
	}, nil
}
