package zone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodePortalZoneWorkflowParams contains parameters for the zone workflow.
type GreenNodePortalZoneWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodePortalZoneWorkflowResult contains the result of the zone workflow.
type GreenNodePortalZoneWorkflowResult struct {
	ZoneCount      int
	DurationMillis int64
}

// GreenNodePortalZoneWorkflow ingests GreenNode zones.
func GreenNodePortalZoneWorkflow(ctx workflow.Context, params GreenNodePortalZoneWorkflowParams) (*GreenNodePortalZoneWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalZoneWorkflow", "region", params.Region)

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

	var result IngestPortalZonesResult
	err := workflow.ExecuteActivity(activityCtx, IngestPortalZonesActivity, IngestPortalZonesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest zones", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodePortalZoneWorkflow", "zoneCount", result.ZoneCount)

	return &GreenNodePortalZoneWorkflowResult{
		ZoneCount:      result.ZoneCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
