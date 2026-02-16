package projectiampolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerProjectIamPolicyWorkflowParams contains parameters for the project IAM policy workflow.
type GCPResourceManagerProjectIamPolicyWorkflowParams struct {
	ProjectID string
}

// GCPResourceManagerProjectIamPolicyWorkflowResult contains the result of the project IAM policy workflow.
type GCPResourceManagerProjectIamPolicyWorkflowResult struct {
	ProjectID   string
	PolicyCount int
}

// GCPResourceManagerProjectIamPolicyWorkflow ingests the GCP project IAM policy for a single project.
func GCPResourceManagerProjectIamPolicyWorkflow(ctx workflow.Context, params GCPResourceManagerProjectIamPolicyWorkflowParams) (*GCPResourceManagerProjectIamPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerProjectIamPolicyWorkflow", "projectID", params.ProjectID)

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

	var result IngestProjectIamPolicyResult
	err := workflow.ExecuteActivity(activityCtx, IngestProjectIamPolicyActivity, IngestProjectIamPolicyParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest project IAM policy", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerProjectIamPolicyWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
	)

	return &GCPResourceManagerProjectIamPolicyWorkflowResult{
		ProjectID:   result.ProjectID,
		PolicyCount: result.PolicyCount,
	}, nil
}
