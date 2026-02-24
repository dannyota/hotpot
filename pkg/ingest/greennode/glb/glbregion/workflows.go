package glbregion

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeGLBGlobalRegionWorkflowParams contains parameters for the region workflow.
type GreenNodeGLBGlobalRegionWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeGLBGlobalRegionWorkflowResult contains the result of the region workflow.
type GreenNodeGLBGlobalRegionWorkflowResult struct {
	RegionCount    int
	DurationMillis int64
}

// GreenNodeGLBGlobalRegionWorkflow ingests GreenNode global regions.
func GreenNodeGLBGlobalRegionWorkflow(ctx workflow.Context, params GreenNodeGLBGlobalRegionWorkflowParams) (*GreenNodeGLBGlobalRegionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeGLBGlobalRegionWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestGLBGlobalRegionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestGLBGlobalRegionsActivity, IngestGLBGlobalRegionsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest global regions", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeGLBGlobalRegionWorkflow", "regionCount", result.RegionCount)

	return &GreenNodeGLBGlobalRegionWorkflowResult{
		RegionCount:    result.RegionCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
