package uptimecheck

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPMonitoringUptimeCheckWorkflowParams contains parameters for the uptime check workflow.
type GCPMonitoringUptimeCheckWorkflowParams struct {
	ProjectID string
}

// GCPMonitoringUptimeCheckWorkflowResult contains the result of the uptime check workflow.
type GCPMonitoringUptimeCheckWorkflowResult struct {
	ProjectID        string
	UptimeCheckCount int
	DurationMillis   int64
}

// GCPMonitoringUptimeCheckWorkflow ingests Monitoring uptime check configs for a single project.
func GCPMonitoringUptimeCheckWorkflow(ctx workflow.Context, params GCPMonitoringUptimeCheckWorkflowParams) (*GCPMonitoringUptimeCheckWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPMonitoringUptimeCheckWorkflow", "projectID", params.ProjectID)

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

	var result IngestUptimeChecksResult
	err := workflow.ExecuteActivity(activityCtx, IngestUptimeChecksActivity, IngestUptimeChecksParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest uptime check configs", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPMonitoringUptimeCheckWorkflow",
		"projectID", params.ProjectID,
		"uptimeCheckCount", result.UptimeCheckCount,
	)

	return &GCPMonitoringUptimeCheckWorkflowResult{
		ProjectID:        result.ProjectID,
		UptimeCheckCount: result.UptimeCheckCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
