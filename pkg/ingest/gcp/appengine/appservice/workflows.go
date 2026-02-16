package appservice

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAppEngineServiceWorkflowParams contains parameters for the service workflow.
type GCPAppEngineServiceWorkflowParams struct {
	ProjectID string
}

// GCPAppEngineServiceWorkflowResult contains the result of the service workflow.
type GCPAppEngineServiceWorkflowResult struct {
	ProjectID      string
	ServiceCount   int
	DurationMillis int64
}

// GCPAppEngineServiceWorkflow ingests App Engine services for a single project.
func GCPAppEngineServiceWorkflow(ctx workflow.Context, params GCPAppEngineServiceWorkflowParams) (*GCPAppEngineServiceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAppEngineServiceWorkflow", "projectID", params.ProjectID)

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

	var result IngestAppEngineServicesResult
	err := workflow.ExecuteActivity(activityCtx, IngestAppEngineServicesActivity, IngestAppEngineServicesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest App Engine services", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAppEngineServiceWorkflow",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
	)

	return &GCPAppEngineServiceWorkflowResult{
		ProjectID:      result.ProjectID,
		ServiceCount:   result.ServiceCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
