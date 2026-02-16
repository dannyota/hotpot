package iampolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPIAPIAMPolicyWorkflowParams contains parameters for the IAP IAM policy workflow.
type GCPIAPIAMPolicyWorkflowParams struct {
	ProjectID string
}

// GCPIAPIAMPolicyWorkflowResult contains the result of the IAP IAM policy workflow.
type GCPIAPIAMPolicyWorkflowResult struct {
	ProjectID      string
	PolicyCount    int
	DurationMillis int64
}

// GCPIAPIAMPolicyWorkflow ingests IAP IAM policy for a single project.
func GCPIAPIAMPolicyWorkflow(ctx workflow.Context, params GCPIAPIAMPolicyWorkflowParams) (*GCPIAPIAMPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAPIAMPolicyWorkflow", "projectID", params.ProjectID)

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

	var result IngestIAPIAMPolicyResult
	err := workflow.ExecuteActivity(activityCtx, IngestIAPIAMPolicyActivity, IngestIAPIAMPolicyParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest IAP IAM policy", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPIAPIAMPolicyWorkflow",
		"projectID", params.ProjectID,
		"policyCount", result.PolicyCount,
	)

	return &GCPIAPIAMPolicyWorkflowResult{
		ProjectID:      result.ProjectID,
		PolicyCount:    result.PolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
