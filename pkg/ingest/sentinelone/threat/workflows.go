package threat

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1ThreatWorkflowResult contains the result of the threat workflow.
type S1ThreatWorkflowResult struct {
	ThreatCount    int
	DurationMillis int64
}

// S1ThreatWorkflow ingests SentinelOne threats.
func S1ThreatWorkflow(ctx workflow.Context) (*S1ThreatWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1ThreatWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestS1ThreatsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1ThreatsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest threats", "error", err)
		return nil, err
	}

	logger.Info("Completed S1ThreatWorkflow", "threatCount", result.ThreatCount)

	return &S1ThreatWorkflowResult{
		ThreatCount:    result.ThreatCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
