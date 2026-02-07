package vpntunnel

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeVpnTunnelWorkflowParams contains parameters for the VPN tunnel workflow.
type GCPComputeVpnTunnelWorkflowParams struct {
	ProjectID string
}

// GCPComputeVpnTunnelWorkflowResult contains the result of the VPN tunnel workflow.
type GCPComputeVpnTunnelWorkflowResult struct {
	ProjectID      string
	VpnTunnelCount int
	DurationMillis int64
}

// GCPComputeVpnTunnelWorkflow ingests GCP Compute VPN tunnels for a single project.
func GCPComputeVpnTunnelWorkflow(ctx workflow.Context, params GCPComputeVpnTunnelWorkflowParams) (*GCPComputeVpnTunnelWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeVpnTunnelWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeVpnTunnelsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeVpnTunnelsActivity, IngestComputeVpnTunnelsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest vpn tunnels", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeVpnTunnelWorkflow",
		"projectID", params.ProjectID,
		"vpnTunnelCount", result.VpnTunnelCount,
	)

	return &GCPComputeVpnTunnelWorkflowResult{
		ProjectID:      result.ProjectID,
		VpnTunnelCount: result.VpnTunnelCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
