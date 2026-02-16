package instance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPSQLInstanceWorkflowParams contains parameters for the SQL instance workflow.
type GCPSQLInstanceWorkflowParams struct {
	ProjectID string
}

// GCPSQLInstanceWorkflowResult contains the result of the SQL instance workflow.
type GCPSQLInstanceWorkflowResult struct {
	ProjectID      string
	InstanceCount  int
	DurationMillis int64
}

// GCPSQLInstanceWorkflow ingests GCP Cloud SQL instances for a single project.
func GCPSQLInstanceWorkflow(ctx workflow.Context, params GCPSQLInstanceWorkflowParams) (*GCPSQLInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSQLInstanceWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestSQLInstancesResult
	err := workflow.ExecuteActivity(activityCtx, IngestSQLInstancesActivity, IngestSQLInstancesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest SQL instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPSQLInstanceWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return &GCPSQLInstanceWorkflowResult{
		ProjectID:      result.ProjectID,
		InstanceCount:  result.InstanceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
