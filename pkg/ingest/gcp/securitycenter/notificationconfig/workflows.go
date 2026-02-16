package notificationconfig

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSecurityCenterNotificationConfigWorkflowParams contains parameters for the SCC notification config workflow.
type GCPSecurityCenterNotificationConfigWorkflowParams struct {
}

// GCPSecurityCenterNotificationConfigWorkflowResult contains the result of the SCC notification config workflow.
type GCPSecurityCenterNotificationConfigWorkflowResult struct {
	NotificationConfigCount int
}

// GCPSecurityCenterNotificationConfigWorkflow ingests SCC notification configs.
func GCPSecurityCenterNotificationConfigWorkflow(ctx workflow.Context, params GCPSecurityCenterNotificationConfigWorkflowParams) (*GCPSecurityCenterNotificationConfigWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSecurityCenterNotificationConfigWorkflow")

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

	var result IngestNotificationConfigsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNotificationConfigsActivity, IngestNotificationConfigsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SCC notification configs", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSecurityCenterNotificationConfigWorkflow",
		"notificationConfigCount", result.NotificationConfigCount,
	)

	return &GCPSecurityCenterNotificationConfigWorkflowResult{
		NotificationConfigCount: result.NotificationConfigCount,
	}, nil
}
