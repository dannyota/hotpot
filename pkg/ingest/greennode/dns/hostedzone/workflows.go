package hostedzone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeDNSHostedZoneWorkflowParams contains parameters for the hosted zone workflow.
type GreenNodeDNSHostedZoneWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeDNSHostedZoneWorkflowResult contains the result of the hosted zone workflow.
type GreenNodeDNSHostedZoneWorkflowResult struct {
	HostedZoneCount int
	DurationMillis  int64
}

// GreenNodeDNSHostedZoneWorkflow ingests GreenNode DNS hosted zones.
func GreenNodeDNSHostedZoneWorkflow(ctx workflow.Context, params GreenNodeDNSHostedZoneWorkflowParams) (*GreenNodeDNSHostedZoneWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeDNSHostedZoneWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestDNSHostedZonesResult
	err := workflow.ExecuteActivity(activityCtx, IngestDNSHostedZonesActivity, IngestDNSHostedZonesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest hosted zones", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeDNSHostedZoneWorkflow",
		"hostedZoneCount", result.HostedZoneCount,
	)

	return &GreenNodeDNSHostedZoneWorkflowResult{
		HostedZoneCount: result.HostedZoneCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
