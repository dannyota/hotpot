package firewall

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeFirewallWorkflowParams contains parameters for the firewall workflow.
type GCPComputeFirewallWorkflowParams struct {
	ProjectID string
}

// GCPComputeFirewallWorkflowResult contains the result of the firewall workflow.
type GCPComputeFirewallWorkflowResult struct {
	ProjectID      string
	FirewallCount  int
	DurationMillis int64
}

// GCPComputeFirewallWorkflow ingests GCP Compute firewalls for a single project.
func GCPComputeFirewallWorkflow(ctx workflow.Context, params GCPComputeFirewallWorkflowParams) (*GCPComputeFirewallWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeFirewallWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeFirewallsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeFirewallsActivity, IngestComputeFirewallsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest firewalls", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeFirewallWorkflow",
		"projectID", params.ProjectID,
		"firewallCount", result.FirewallCount,
	)

	return &GCPComputeFirewallWorkflowResult{
		ProjectID:      result.ProjectID,
		FirewallCount:  result.FirewallCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
