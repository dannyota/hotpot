package service

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPRunServiceWorkflowParams contains parameters for the Cloud Run service workflow.
type GCPRunServiceWorkflowParams struct {
	ProjectID string
}

// GCPRunServiceWorkflowResult contains the result of the Cloud Run service workflow.
type GCPRunServiceWorkflowResult struct {
	ProjectID    string
	ServiceCount int
}

// GCPRunServiceWorkflow ingests Cloud Run services for a single project.
func GCPRunServiceWorkflow(ctx workflow.Context, params GCPRunServiceWorkflowParams) (*GCPRunServiceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPRunServiceWorkflow", "projectID", params.ProjectID)

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

	var result IngestRunServicesResult
	err := workflow.ExecuteActivity(activityCtx, IngestRunServicesActivity, IngestRunServicesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Cloud Run services", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPRunServiceWorkflow",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
	)

	return &GCPRunServiceWorkflowResult{
		ProjectID:    result.ProjectID,
		ServiceCount: result.ServiceCount,
	}, nil
}
