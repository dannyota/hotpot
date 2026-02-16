package dnspolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPDNSPolicyWorkflowParams contains parameters for the DNS policy workflow.
type GCPDNSPolicyWorkflowParams struct {
	ProjectID string
}

// GCPDNSPolicyWorkflowResult contains the result of the DNS policy workflow.
type GCPDNSPolicyWorkflowResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// GCPDNSPolicyWorkflow ingests GCP DNS policies for a single project.
func GCPDNSPolicyWorkflow(ctx workflow.Context, params GCPDNSPolicyWorkflowParams) (*GCPDNSPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPDNSPolicyWorkflow", "projectID", params.ProjectID)

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
	var result IngestDNSPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestDNSPoliciesActivity, IngestDNSPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPDNSPolicyWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
	)

	return &GCPDNSPolicyWorkflowResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
