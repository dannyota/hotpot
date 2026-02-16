package router

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeRouterWorkflowParams contains parameters for the router workflow.
type GCPComputeRouterWorkflowParams struct {
	ProjectID string
}

// GCPComputeRouterWorkflowResult contains the result of the router workflow.
type GCPComputeRouterWorkflowResult struct {
	ProjectID      string
	RouterCount    int
	DurationMillis int64
}

// GCPComputeRouterWorkflow ingests GCP Compute routers for a single project.
func GCPComputeRouterWorkflow(ctx workflow.Context, params GCPComputeRouterWorkflowParams) (*GCPComputeRouterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeRouterWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeRoutersResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeRoutersActivity, IngestComputeRoutersParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest routers", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeRouterWorkflow",
		"projectID", params.ProjectID,
		"routerCount", result.RouterCount,
	)

	return &GCPComputeRouterWorkflowResult{
		ProjectID:      result.ProjectID,
		RouterCount:    result.RouterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
