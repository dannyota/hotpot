package serviceaccountkey

import (
	"time"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPIAMServiceAccountKeyWorkflowParams struct {
	ProjectID      string
	QuotaProjectID string
}

type GCPIAMServiceAccountKeyWorkflowResult struct {
	ProjectID              string
	ServiceAccountKeyCount int
	DurationMillis         int64
}

func GCPIAMServiceAccountKeyWorkflow(ctx workflow.Context, params GCPIAMServiceAccountKeyWorkflowParams) (*GCPIAMServiceAccountKeyWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAMServiceAccountKeyWorkflow", "projectID", params.ProjectID)

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

	var result IngestIAMServiceAccountKeysResult
	err := workflow.ExecuteActivity(activityCtx, IngestIAMServiceAccountKeysActivity, IngestIAMServiceAccountKeysParams{
		ProjectID:      params.ProjectID,
		QuotaProjectID: params.QuotaProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest service account keys", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GCPIAMServiceAccountKeyWorkflow",
		"projectID", params.ProjectID,
		"serviceAccountKeyCount", result.ServiceAccountKeyCount,
	)

	return &GCPIAMServiceAccountKeyWorkflowResult{
		ProjectID:              result.ProjectID,
		ServiceAccountKeyCount: result.ServiceAccountKeyCount,
		DurationMillis:         result.DurationMillis,
	}, nil
}
