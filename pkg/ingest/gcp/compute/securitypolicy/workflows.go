package securitypolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeSecurityPolicyWorkflowParams contains parameters for the security policy workflow.
type GCPComputeSecurityPolicyWorkflowParams struct {
	ProjectID string
}

// GCPComputeSecurityPolicyWorkflowResult contains the result of the security policy workflow.
type GCPComputeSecurityPolicyWorkflowResult struct {
	ProjectID           string
	SecurityPolicyCount int
	DurationMillis      int64
}

// GCPComputeSecurityPolicyWorkflow ingests GCP Compute security policies for a single project.
func GCPComputeSecurityPolicyWorkflow(ctx workflow.Context, params GCPComputeSecurityPolicyWorkflowParams) (*GCPComputeSecurityPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSecurityPolicyWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeSecurityPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeSecurityPoliciesActivity, IngestComputeSecurityPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest security policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeSecurityPolicyWorkflow",
		"projectID", params.ProjectID,
		"securityPolicyCount", result.SecurityPolicyCount,
	)

	return &GCPComputeSecurityPolicyWorkflowResult{
		ProjectID:           result.ProjectID,
		SecurityPolicyCount: result.SecurityPolicyCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
