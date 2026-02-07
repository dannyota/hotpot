package subnetwork

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeSubnetworkWorkflowParams contains parameters for the subnetwork workflow.
type GCPComputeSubnetworkWorkflowParams struct {
	ProjectID string
}

// GCPComputeSubnetworkWorkflowResult contains the result of the subnetwork workflow.
type GCPComputeSubnetworkWorkflowResult struct {
	ProjectID       string
	SubnetworkCount int
	DurationMillis  int64
}

// GCPComputeSubnetworkWorkflow ingests GCP Compute subnetworks for a single project.
func GCPComputeSubnetworkWorkflow(ctx workflow.Context, params GCPComputeSubnetworkWorkflowParams) (*GCPComputeSubnetworkWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSubnetworkWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestComputeSubnetworksResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeSubnetworksActivity, IngestComputeSubnetworksParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest subnetworks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeSubnetworkWorkflow",
		"projectID", params.ProjectID,
		"subnetworkCount", result.SubnetworkCount,
	)

	return &GCPComputeSubnetworkWorkflowResult{
		ProjectID:       result.ProjectID,
		SubnetworkCount: result.SubnetworkCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
