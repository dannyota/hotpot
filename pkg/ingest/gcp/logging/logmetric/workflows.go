package logmetric

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPLoggingLogMetricWorkflowParams contains parameters for the log metric workflow.
type GCPLoggingLogMetricWorkflowParams struct {
	ProjectID string
}

// GCPLoggingLogMetricWorkflowResult contains the result of the log metric workflow.
type GCPLoggingLogMetricWorkflowResult struct {
	ProjectID      string
	LogMetricCount int
	DurationMillis int64
}

// GCPLoggingLogMetricWorkflow ingests GCP Cloud Logging log metrics for a single project.
func GCPLoggingLogMetricWorkflow(ctx workflow.Context, params GCPLoggingLogMetricWorkflowParams) (*GCPLoggingLogMetricWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPLoggingLogMetricWorkflow", "projectID", params.ProjectID)

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

	var result IngestLoggingLogMetricsResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoggingLogMetricsActivity, IngestLoggingLogMetricsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest log metrics", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPLoggingLogMetricWorkflow",
		"projectID", params.ProjectID,
		"logMetricCount", result.LogMetricCount,
	)

	return &GCPLoggingLogMetricWorkflowResult{
		ProjectID:      result.ProjectID,
		LogMetricCount: result.LogMetricCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
