package quota

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodePortalQuotaWorkflowResult contains the result of the quota workflow.
type GreenNodePortalQuotaWorkflowResult struct {
	QuotaCount     int
	DurationMillis int64
}

// GreenNodePortalQuotaWorkflow ingests GreenNode quotas.
func GreenNodePortalQuotaWorkflow(ctx workflow.Context) (*GreenNodePortalQuotaWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalQuotaWorkflow")

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
	err := workflow.ExecuteActivity(activityCtx, IngestPortalQuotasActivity, IngestPortalQuotasParams{}).Get(ctx, &result)
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
