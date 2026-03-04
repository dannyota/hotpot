package eol

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
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

	var result IngestEOLResult
	err := workflow.ExecuteActivity(activityCtx, IngestEOLActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest EOL data", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed EOLWorkflow",
		"productCount", result.ProductCount,
		"cycleCount", result.CycleCount,
	)

	return &EOLWorkflowResult{
		ProductCount:   result.ProductCount,
		CycleCount:     result.CycleCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
