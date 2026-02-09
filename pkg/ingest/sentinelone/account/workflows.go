package account

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1AccountWorkflowResult contains the result of the account workflow.
type S1AccountWorkflowResult struct {
	AccountCount   int
	DurationMillis int64
}

// S1AccountWorkflow ingests SentinelOne accounts.
func S1AccountWorkflow(ctx workflow.Context) (*S1AccountWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1AccountWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestS1AccountsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1AccountsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest accounts", "error", err)
		return nil, err
	}

	logger.Info("Completed S1AccountWorkflow", "accountCount", result.AccountCount)

	return &S1AccountWorkflowResult{
		AccountCount:   result.AccountCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
