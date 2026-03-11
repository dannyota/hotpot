package subnet

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkSubnetWorkflowParams contains parameters for the subnet workflow.
type GreenNodeNetworkSubnetWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkSubnetWorkflowResult contains the result of the subnet workflow.
type GreenNodeNetworkSubnetWorkflowResult struct {
	SubnetCount    int
	DurationMillis int64
}

// GreenNodeNetworkSubnetWorkflow ingests GreenNode subnets.
func GreenNodeNetworkSubnetWorkflow(ctx workflow.Context, params GreenNodeNetworkSubnetWorkflowParams) (*GreenNodeNetworkSubnetWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkSubnetWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkSubnetsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkSubnetsActivity, IngestNetworkSubnetsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest subnets", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkSubnetWorkflow", "subnetCount", result.SubnetCount)

	return &GreenNodeNetworkSubnetWorkflowResult{
		SubnetCount:    result.SubnetCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
