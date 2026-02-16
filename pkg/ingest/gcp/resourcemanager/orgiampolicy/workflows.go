package orgiampolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerOrgIamPolicyWorkflowParams contains parameters for the org IAM policy workflow.
type GCPResourceManagerOrgIamPolicyWorkflowParams struct {
}

// GCPResourceManagerOrgIamPolicyWorkflowResult contains the result of the org IAM policy workflow.
type GCPResourceManagerOrgIamPolicyWorkflowResult struct {
	PolicyCount int
}

// GCPResourceManagerOrgIamPolicyWorkflow ingests GCP organization IAM policies.
func GCPResourceManagerOrgIamPolicyWorkflow(ctx workflow.Context, params GCPResourceManagerOrgIamPolicyWorkflowParams) (*GCPResourceManagerOrgIamPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerOrgIamPolicyWorkflow")

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

	var result IngestOrgIamPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestOrgIamPoliciesActivity, IngestOrgIamPoliciesParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest org IAM policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerOrgIamPolicyWorkflow",
		"policyCount", result.PolicyCount,
	)

	return &GCPResourceManagerOrgIamPolicyWorkflowResult{
		PolicyCount: result.PolicyCount,
	}, nil
}
