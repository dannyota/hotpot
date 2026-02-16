package database

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSpannerDatabaseWorkflowParams contains parameters for the Spanner database workflow.
type GCPSpannerDatabaseWorkflowParams struct {
	ProjectID     string
	InstanceNames []string
}

// GCPSpannerDatabaseWorkflowResult contains the result of the Spanner database workflow.
type GCPSpannerDatabaseWorkflowResult struct {
	ProjectID      string
	DatabaseCount  int
	DurationMillis int64
}

// GCPSpannerDatabaseWorkflow ingests GCP Spanner databases for a single project.
func GCPSpannerDatabaseWorkflow(ctx workflow.Context, params GCPSpannerDatabaseWorkflowParams) (*GCPSpannerDatabaseWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSpannerDatabaseWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", len(params.InstanceNames),
	)

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

	var result IngestSpannerDatabasesResult
	err := workflow.ExecuteActivity(activityCtx, IngestSpannerDatabasesActivity, IngestSpannerDatabasesParams{
		ProjectID:     params.ProjectID,
		InstanceNames: params.InstanceNames,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Spanner databases", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSpannerDatabaseWorkflow",
		"projectID", params.ProjectID,
		"databaseCount", result.DatabaseCount,
	)

	return &GCPSpannerDatabaseWorkflowResult{
		ProjectID:      result.ProjectID,
		DatabaseCount:  result.DatabaseCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
