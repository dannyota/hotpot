package quota

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodePortalQuotaWorkflowParams contains parameters for the quota workflow.
type GreenNodePortalQuotaWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodePortalQuotaWorkflowResult contains the result of the quota workflow.
type GreenNodePortalQuotaWorkflowResult struct {
	QuotaCount     int
	DurationMillis int64
}

// GreenNodePortalQuotaWorkflow ingests GreenNode quotas.
func GreenNodePortalQuotaWorkflow(ctx workflow.Context, params GreenNodePortalQuotaWorkflowParams) (*GreenNodePortalQuotaWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalQuotaWorkflow", "region", params.Region)

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

	var result IngestPortalQuotasResult
	err := workflow.ExecuteActivity(activityCtx, IngestPortalQuotasActivity, IngestPortalQuotasParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest quotas", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodePortalQuotaWorkflow", "quotaCount", result.QuotaCount)

	return &GreenNodePortalQuotaWorkflowResult{
		QuotaCount:     result.QuotaCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
