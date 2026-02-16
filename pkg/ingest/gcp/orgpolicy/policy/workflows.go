package policy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPOrgPolicyPolicyWorkflowParams struct {
}

type GCPOrgPolicyPolicyWorkflowResult struct {
	PolicyCount int
}

func GCPOrgPolicyPolicyWorkflow(ctx workflow.Context, params GCPOrgPolicyPolicyWorkflowParams) (*GCPOrgPolicyPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPOrgPolicyPolicyWorkflow")

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

	var result IngestPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestPoliciesActivity, IngestPoliciesParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest org policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPOrgPolicyPolicyWorkflow",
		"policyCount", result.PolicyCount,
	)

	return &GCPOrgPolicyPolicyWorkflowResult{
		PolicyCount: result.PolicyCount,
	}, nil
}
