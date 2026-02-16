package secretmanager

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager/secret"
)

// GCPSecretManagerWorkflowParams contains parameters for the Secret Manager workflow.
type GCPSecretManagerWorkflowParams struct {
	ProjectID string
}

// GCPSecretManagerWorkflowResult contains the result of the Secret Manager workflow.
type GCPSecretManagerWorkflowResult struct {
	ProjectID   string
	SecretCount int
}

// GCPSecretManagerWorkflow ingests all GCP Secret Manager resources for a single project.
func GCPSecretManagerWorkflow(ctx workflow.Context, params GCPSecretManagerWorkflowParams) (*GCPSecretManagerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecretManagerWorkflow", "projectID", params.ProjectID)

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

	result := &GCPSecretManagerWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var secretResult secret.GCPSecretManagerSecretWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, secret.GCPSecretManagerSecretWorkflow,
		secret.GCPSecretManagerSecretWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &secretResult)
	if err != nil {
		logger.Error("Failed to ingest secrets", "error", err)
		return nil, err
	}
	result.SecretCount = secretResult.SecretCount

	logger.Info("Completed GCPSecretManagerWorkflow",
		"projectID", params.ProjectID,
		"secretCount", result.SecretCount,
	)

	return result, nil
}
