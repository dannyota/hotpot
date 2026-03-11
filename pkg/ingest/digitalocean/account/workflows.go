package account

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOAccountWorkflowResult contains the result of the Account workflow.
type DOAccountWorkflowResult struct {
	AccountCount   int
	DurationMillis int64
}

// DOAccountWorkflow ingests DigitalOcean Account.
func DOAccountWorkflow(ctx workflow.Context) (*DOAccountWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOAccountWorkflow")

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

	var result IngestDOAccountsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOAccountsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Account", "error", err)
		return nil, err
	}

	logger.Info("Completed DOAccountWorkflow", "accountCount", result.AccountCount)

	return &DOAccountWorkflowResult{
		AccountCount:   result.AccountCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
