package key

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOKeyWorkflowResult contains the result of the SSH key workflow.
type DOKeyWorkflowResult struct {
	KeyCount       int
	DurationMillis int64
}

// DOKeyWorkflow ingests DigitalOcean SSH keys.
func DOKeyWorkflow(ctx workflow.Context) (*DOKeyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOKeyWorkflow")

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

	var result IngestDOKeysResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOKeysActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest keys", "error", err)
		return nil, err
	}

	logger.Info("Completed DOKeyWorkflow", "keyCount", result.KeyCount)

	return &DOKeyWorkflowResult{
		KeyCount:       result.KeyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
