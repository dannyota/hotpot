package vpngateway

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeVpnGatewayWorkflowParams contains parameters for the VPN gateway workflow.
type GCPComputeVpnGatewayWorkflowParams struct {
	ProjectID string
}

// GCPComputeVpnGatewayWorkflowResult contains the result of the VPN gateway workflow.
type GCPComputeVpnGatewayWorkflowResult struct {
	ProjectID       string
	VpnGatewayCount int
	DurationMillis  int64
}

// GCPComputeVpnGatewayWorkflow ingests GCP Compute VPN gateways for a single project.
func GCPComputeVpnGatewayWorkflow(ctx workflow.Context, params GCPComputeVpnGatewayWorkflowParams) (*GCPComputeVpnGatewayWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeVpnGatewayWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeVpnGatewaysResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeVpnGatewaysActivity, IngestComputeVpnGatewaysParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest vpn gateways", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeVpnGatewayWorkflow",
		"projectID", params.ProjectID,
		"vpnGatewayCount", result.VpnGatewayCount,
	)

	return &GCPComputeVpnGatewayWorkflowResult{
		ProjectID:       result.ProjectID,
		VpnGatewayCount: result.VpnGatewayCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
