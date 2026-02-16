package folderiampolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerFolderIamPolicyWorkflowParams contains parameters for the folder IAM policy workflow.
type GCPResourceManagerFolderIamPolicyWorkflowParams struct {
}

// GCPResourceManagerFolderIamPolicyWorkflowResult contains the result of the folder IAM policy workflow.
type GCPResourceManagerFolderIamPolicyWorkflowResult struct {
	PolicyCount int
}

// GCPResourceManagerFolderIamPolicyWorkflow ingests GCP folder IAM policies.
func GCPResourceManagerFolderIamPolicyWorkflow(ctx workflow.Context, params GCPResourceManagerFolderIamPolicyWorkflowParams) (*GCPResourceManagerFolderIamPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerFolderIamPolicyWorkflow")

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

	var result IngestResourceManagerFolderIamPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestResourceManagerFolderIamPoliciesActivity, IngestResourceManagerFolderIamPoliciesParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest folder IAM policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerFolderIamPolicyWorkflow",
		"policyCount", result.PolicyCount,
	)

	return &GCPResourceManagerFolderIamPolicyWorkflowResult{
		PolicyCount: result.PolicyCount,
	}, nil
}
