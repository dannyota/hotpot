package targetvpngateway

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetVpnGatewayWorkflowParams contains parameters for the target VPN gateway workflow.
type GCPComputeTargetVpnGatewayWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetVpnGatewayWorkflowResult contains the result of the target VPN gateway workflow.
type GCPComputeTargetVpnGatewayWorkflowResult struct {
	ProjectID             string
	TargetVpnGatewayCount int
	DurationMillis        int64
}

// GCPComputeTargetVpnGatewayWorkflow ingests GCP Compute Classic VPN gateways for a single project.
func GCPComputeTargetVpnGatewayWorkflow(ctx workflow.Context, params GCPComputeTargetVpnGatewayWorkflowParams) (*GCPComputeTargetVpnGatewayWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetVpnGatewayWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeTargetVpnGatewaysResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeTargetVpnGatewaysActivity, IngestComputeTargetVpnGatewaysParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target vpn gateways", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetVpnGatewayWorkflow",
		"projectID", params.ProjectID,
		"targetVpnGatewayCount", result.TargetVpnGatewayCount,
	)

	return &GCPComputeTargetVpnGatewayWorkflowResult{
		ProjectID:             result.ProjectID,
		TargetVpnGatewayCount: result.TargetVpnGatewayCount,
		DurationMillis:        result.DurationMillis,
	}, nil
}
