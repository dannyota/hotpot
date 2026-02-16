package application

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAppEngineApplicationWorkflowParams contains parameters for the application workflow.
type GCPAppEngineApplicationWorkflowParams struct {
	ProjectID string
}

// GCPAppEngineApplicationWorkflowResult contains the result of the application workflow.
type GCPAppEngineApplicationWorkflowResult struct {
	ProjectID        string
	ApplicationCount int
	DurationMillis   int64
}

// GCPAppEngineApplicationWorkflow ingests App Engine applications for a single project.
func GCPAppEngineApplicationWorkflow(ctx workflow.Context, params GCPAppEngineApplicationWorkflowParams) (*GCPAppEngineApplicationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAppEngineApplicationWorkflow", "projectID", params.ProjectID)

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

	var result IngestAppEngineApplicationsResult
	err := workflow.ExecuteActivity(activityCtx, IngestAppEngineApplicationsActivity, IngestAppEngineApplicationsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest App Engine applications", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAppEngineApplicationWorkflow",
		"projectID", params.ProjectID,
		"applicationCount", result.ApplicationCount,
	)

	return &GCPAppEngineApplicationWorkflowResult{
		ProjectID:        result.ProjectID,
		ApplicationCount: result.ApplicationCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
