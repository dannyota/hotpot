package iam

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
)

// GCPIAMWorkflowParams contains parameters for the IAM workflow.
type GCPIAMWorkflowParams struct {
	ProjectID string
}

// GCPIAMWorkflowResult contains the result of the IAM workflow.
type GCPIAMWorkflowResult struct {
	ProjectID              string
	ServiceAccountCount    int
	ServiceAccountKeyCount int
}

// GCPIAMWorkflow ingests all IAM resources for a single project.
func GCPIAMWorkflow(ctx workflow.Context, params GCPIAMWorkflowParams) (*GCPIAMWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAMWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPIAMWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute service account workflow
	var saResult serviceaccount.GCPIAMServiceAccountWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, serviceaccount.GCPIAMServiceAccountWorkflow,
		serviceaccount.GCPIAMServiceAccountWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &saResult)
	if err != nil {
		logger.Error("Failed to ingest service accounts", "error", err)
		return nil, err
	}
	result.ServiceAccountCount = saResult.ServiceAccountCount

	// Execute service account key workflow
	var keyResult serviceaccountkey.GCPIAMServiceAccountKeyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, serviceaccountkey.GCPIAMServiceAccountKeyWorkflow,
		serviceaccountkey.GCPIAMServiceAccountKeyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &keyResult)
	if err != nil {
		logger.Error("Failed to ingest service account keys", "error", err)
		return nil, err
	}
	result.ServiceAccountKeyCount = keyResult.ServiceAccountKeyCount

	logger.Info("Completed GCPIAMWorkflow",
		"projectID", params.ProjectID,
		"serviceAccountCount", result.ServiceAccountCount,
		"serviceAccountKeyCount", result.ServiceAccountKeyCount,
	)

	return result, nil
}
