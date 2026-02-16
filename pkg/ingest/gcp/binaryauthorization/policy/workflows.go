package policy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBinaryAuthorizationPolicyWorkflowParams contains parameters for the policy workflow.
type GCPBinaryAuthorizationPolicyWorkflowParams struct {
	ProjectID string
}

// GCPBinaryAuthorizationPolicyWorkflowResult contains the result of the policy workflow.
type GCPBinaryAuthorizationPolicyWorkflowResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// GCPBinaryAuthorizationPolicyWorkflow ingests Binary Authorization policies for a single project.
func GCPBinaryAuthorizationPolicyWorkflow(ctx workflow.Context, params GCPBinaryAuthorizationPolicyWorkflowParams) (*GCPBinaryAuthorizationPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBinaryAuthorizationPolicyWorkflow", "projectID", params.ProjectID)

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

	var result IngestBinaryAuthorizationPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestBinaryAuthorizationPoliciesActivity, IngestBinaryAuthorizationPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest binary authorization policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBinaryAuthorizationPolicyWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
	)

	return &GCPBinaryAuthorizationPolicyWorkflowResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
