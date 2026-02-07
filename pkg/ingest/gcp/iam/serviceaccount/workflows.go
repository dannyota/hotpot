package serviceaccount

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPIAMServiceAccountWorkflowParams struct {
	ProjectID string
}

type GCPIAMServiceAccountWorkflowResult struct {
	ProjectID           string
	ServiceAccountCount int
	DurationMillis      int64
}

func GCPIAMServiceAccountWorkflow(ctx workflow.Context, params GCPIAMServiceAccountWorkflowParams) (*GCPIAMServiceAccountWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAMServiceAccountWorkflow", "projectID", params.ProjectID)

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

	var result IngestIAMServiceAccountsResult
	err := workflow.ExecuteActivity(activityCtx, IngestIAMServiceAccountsActivity, IngestIAMServiceAccountsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest service accounts", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPIAMServiceAccountWorkflow",
		"projectID", params.ProjectID,
		"serviceAccountCount", result.ServiceAccountCount,
	)

	return &GCPIAMServiceAccountWorkflowResult{
		ProjectID:           result.ProjectID,
		ServiceAccountCount: result.ServiceAccountCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
