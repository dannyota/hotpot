package settings

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPIAPSettingsWorkflowParams contains parameters for the IAP settings workflow.
type GCPIAPSettingsWorkflowParams struct {
	ProjectID string
}

// GCPIAPSettingsWorkflowResult contains the result of the IAP settings workflow.
type GCPIAPSettingsWorkflowResult struct {
	ProjectID      string
	SettingsCount  int
	DurationMillis int64
}

// GCPIAPSettingsWorkflow ingests IAP settings for a single project.
func GCPIAPSettingsWorkflow(ctx workflow.Context, params GCPIAPSettingsWorkflowParams) (*GCPIAPSettingsWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPIAPSettingsWorkflow", "projectID", params.ProjectID)

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

	var result IngestIAPSettingsResult
	err := workflow.ExecuteActivity(activityCtx, IngestIAPSettingsActivity, IngestIAPSettingsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest IAP settings", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPIAPSettingsWorkflow",
		"projectID", params.ProjectID,
		"settingsCount", result.SettingsCount,
	)

	return &GCPIAPSettingsWorkflowResult{
		ProjectID:      result.ProjectID,
		SettingsCount:  result.SettingsCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
