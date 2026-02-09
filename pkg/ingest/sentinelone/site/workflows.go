package site

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1SiteWorkflowResult contains the result of the site workflow.
type S1SiteWorkflowResult struct {
	SiteCount      int
	DurationMillis int64
}

// S1SiteWorkflow ingests SentinelOne sites.
func S1SiteWorkflow(ctx workflow.Context) (*S1SiteWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1SiteWorkflow")

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

	var result IngestS1SitesResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1SitesActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest sites", "error", err)
		return nil, err
	}

	logger.Info("Completed S1SiteWorkflow", "siteCount", result.SiteCount)

	return &S1SiteWorkflowResult{
		SiteCount:      result.SiteCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
