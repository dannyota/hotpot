package peering

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkPeeringWorkflowParams contains parameters for the peering workflow.
type GreenNodeNetworkPeeringWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkPeeringWorkflowResult contains the result of the peering workflow.
type GreenNodeNetworkPeeringWorkflowResult struct {
	PeeringCount   int
	DurationMillis int64
}

// GreenNodeNetworkPeeringWorkflow ingests GreenNode peerings.
func GreenNodeNetworkPeeringWorkflow(ctx workflow.Context, params GreenNodeNetworkPeeringWorkflowParams) (*GreenNodeNetworkPeeringWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkPeeringWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkPeeringsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkPeeringsActivity, IngestNetworkPeeringsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest peerings", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkPeeringWorkflow", "peeringCount", result.PeeringCount)

	return &GreenNodeNetworkPeeringWorkflowResult{
		PeeringCount:   result.PeeringCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
