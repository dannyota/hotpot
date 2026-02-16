package sink

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPLoggingSinkWorkflowParams contains parameters for the sink workflow.
type GCPLoggingSinkWorkflowParams struct {
	ProjectID string
}

// GCPLoggingSinkWorkflowResult contains the result of the sink workflow.
type GCPLoggingSinkWorkflowResult struct {
	ProjectID      string
	SinkCount      int
	DurationMillis int64
}

// GCPLoggingSinkWorkflow ingests GCP Cloud Logging sinks for a single project.
func GCPLoggingSinkWorkflow(ctx workflow.Context, params GCPLoggingSinkWorkflowParams) (*GCPLoggingSinkWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPLoggingSinkWorkflow", "projectID", params.ProjectID)

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

	var result IngestLoggingSinksResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoggingSinksActivity, IngestLoggingSinksParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest sinks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPLoggingSinkWorkflow",
		"projectID", params.ProjectID,
		"sinkCount", result.SinkCount,
	)

	return &GCPLoggingSinkWorkflowResult{
		ProjectID:      result.ProjectID,
		SinkCount:      result.SinkCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
