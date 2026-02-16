package secret

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSecretManagerSecretWorkflowParams contains parameters for the secret workflow.
type GCPSecretManagerSecretWorkflowParams struct {
	ProjectID string
}

// GCPSecretManagerSecretWorkflowResult contains the result of the secret workflow.
type GCPSecretManagerSecretWorkflowResult struct {
	ProjectID      string
	SecretCount    int
	DurationMillis int64
}

// GCPSecretManagerSecretWorkflow ingests GCP Secret Manager secrets for a single project.
func GCPSecretManagerSecretWorkflow(ctx workflow.Context, params GCPSecretManagerSecretWorkflowParams) (*GCPSecretManagerSecretWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecretManagerSecretWorkflow", "projectID", params.ProjectID)

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

	var result IngestSecretManagerSecretsResult
	err := workflow.ExecuteActivity(activityCtx, IngestSecretManagerSecretsActivity, IngestSecretManagerSecretsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest secrets", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSecretManagerSecretWorkflow",
		"projectID", params.ProjectID,
		"secretCount", result.SecretCount,
	)

	return &GCPSecretManagerSecretWorkflowResult{
		ProjectID:      result.ProjectID,
		SecretCount:    result.SecretCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
