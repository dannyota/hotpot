package droplet

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DODropletWorkflowResult contains the result of the Droplet workflow.
type DODropletWorkflowResult struct {
	DropletCount   int
	DurationMillis int64
}

// DODropletWorkflow ingests DigitalOcean Droplets.
func DODropletWorkflow(ctx workflow.Context) (*DODropletWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DODropletWorkflow")

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

	var result IngestDODropletsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDODropletsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest droplets", "error", err)
		return nil, err
	}

	logger.Info("Completed DODropletWorkflow", "dropletCount", result.DropletCount)

	return &DODropletWorkflowResult{
		DropletCount:   result.DropletCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
