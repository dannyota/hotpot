package sslpolicy

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeSslPolicyWorkflowParams contains parameters for the SSL policy workflow.
type GCPComputeSslPolicyWorkflowParams struct {
	ProjectID string
}

// GCPComputeSslPolicyWorkflowResult contains the result of the SSL policy workflow.
type GCPComputeSslPolicyWorkflowResult struct {
	ProjectID      string
	SslPolicyCount int
	DurationMillis int64
}

// GCPComputeSslPolicyWorkflow ingests GCP Compute SSL policies for a single project.
func GCPComputeSslPolicyWorkflow(ctx workflow.Context, params GCPComputeSslPolicyWorkflowParams) (*GCPComputeSslPolicyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSslPolicyWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeSslPoliciesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeSslPoliciesActivity, IngestComputeSslPoliciesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SSL policies", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeSslPolicyWorkflow",
		"projectID", params.ProjectID,
		"sslPolicyCount", result.SslPolicyCount,
	)

	return &GCPComputeSslPolicyWorkflowResult{
		ProjectID:      result.ProjectID,
		SslPolicyCount: result.SslPolicyCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
