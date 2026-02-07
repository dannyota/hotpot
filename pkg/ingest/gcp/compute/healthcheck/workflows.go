package healthcheck

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeHealthCheckWorkflowParams contains parameters for the health check workflow.
type GCPComputeHealthCheckWorkflowParams struct {
	ProjectID string
}

// GCPComputeHealthCheckWorkflowResult contains the result of the health check workflow.
type GCPComputeHealthCheckWorkflowResult struct {
	ProjectID        string
	HealthCheckCount int
	DurationMillis   int64
}

// GCPComputeHealthCheckWorkflow ingests GCP Compute health checks for a single project.
func GCPComputeHealthCheckWorkflow(ctx workflow.Context, params GCPComputeHealthCheckWorkflowParams) (*GCPComputeHealthCheckWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeHealthCheckWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeHealthChecksResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeHealthChecksActivity, IngestComputeHealthChecksParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest health checks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeHealthCheckWorkflow",
		"projectID", params.ProjectID,
		"healthCheckCount", result.HealthCheckCount,
	)

	return &GCPComputeHealthCheckWorkflowResult{
		ProjectID:        result.ProjectID,
		HealthCheckCount: result.HealthCheckCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
