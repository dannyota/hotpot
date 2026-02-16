package accesspolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAccessContextManagerAccessPolicyWorkflowParams contains parameters for the access policy workflow.
type GCPAccessContextManagerAccessPolicyWorkflowParams struct {
}

// GCPAccessContextManagerAccessPolicyWorkflowResult contains the result of the access policy workflow.
type GCPAccessContextManagerAccessPolicyWorkflowResult struct {
	PolicyCount int
}

// GCPAccessContextManagerAccessPolicyWorkflow ingests access policies.
func GCPAccessContextManagerAccessPolicyWorkflow(ctx workflow.Context, params GCPAccessContextManagerAccessPolicyWorkflowParams) (*GCPAccessContextManagerAccessPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAccessContextManagerAccessPolicyWorkflow")

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

	var result IngestAccessPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestAccessPoliciesActivity, IngestAccessPoliciesParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest access policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAccessContextManagerAccessPolicyWorkflow",
		"policyCount", result.PolicyCount,
	)

	return &GCPAccessContextManagerAccessPolicyWorkflowResult{
		PolicyCount: result.PolicyCount,
	}, nil
}
