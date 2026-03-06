package cpe

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// CPEWorkflowResult contains the result of the CPE workflow.
type CPEWorkflowResult struct {
	CPECount       int
	DurationMillis int64
}

// CPEWorkflow ingests the NVD CPE Dictionary.
func CPEWorkflow(ctx workflow.Context) (*CPEWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting CPEWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Minute,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestCPEResult
	err := workflow.ExecuteActivity(activityCtx, IngestCPEActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest CPE data", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed CPEWorkflow", "cpeCount", result.CPECount)

	return &CPEWorkflowResult{
		CPECount:       result.CPECount,
		DurationMillis: result.DurationMillis,
	}, nil
}
