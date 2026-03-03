package computer

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// MEECComputerWorkflowResult contains the result of the computer workflow.
type MEECComputerWorkflowResult struct {
	ComputerCount  int
	DurationMillis int64
}

// MEECComputerWorkflow ingests MEEC computers.
func MEECComputerWorkflow(ctx workflow.Context) (*MEECComputerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting MEECComputerWorkflow")

	// Step 1: Ingest computers
	ingestOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ingestCtx := workflow.WithActivityOptions(ctx, ingestOpts)

	var ingestResult IngestComputersResult
	err := workflow.ExecuteActivity(ingestCtx, IngestComputersActivity).Get(ctx, &ingestResult)
	if err != nil {
		logger.Error("Failed to ingest computers", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed MEECComputerWorkflow", "computerCount", ingestResult.ComputerCount)

	return &MEECComputerWorkflowResult{
		ComputerCount:  ingestResult.ComputerCount,
		DurationMillis: ingestResult.DurationMillis,
	}, nil
}
