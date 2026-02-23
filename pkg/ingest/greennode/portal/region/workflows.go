package region

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodePortalRegionWorkflowParams contains parameters for the region workflow.
type GreenNodePortalRegionWorkflowParams struct {
	Region string
}

// GreenNodePortalRegionWorkflowResult contains the result of the region workflow.
type GreenNodePortalRegionWorkflowResult struct {
	RegionCount    int
	DurationMillis int64
}

// GreenNodePortalRegionWorkflow ingests GreenNode regions.
func GreenNodePortalRegionWorkflow(ctx workflow.Context, params GreenNodePortalRegionWorkflowParams) (*GreenNodePortalRegionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalRegionWorkflow", "region", params.Region)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestPortalRegionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestPortalRegionsActivity, IngestPortalRegionsParams{
		Region: params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest regions", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodePortalRegionWorkflow", "regionCount", result.RegionCount)

	return &GreenNodePortalRegionWorkflowResult{
		RegionCount:    result.RegionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
