package backendservice

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeBackendServiceWorkflowParams contains parameters for the backend service workflow.
type GCPComputeBackendServiceWorkflowParams struct {
	ProjectID string
}

// GCPComputeBackendServiceWorkflowResult contains the result of the backend service workflow.
type GCPComputeBackendServiceWorkflowResult struct {
	ProjectID           string
	BackendServiceCount int
	DurationMillis      int64
}

// GCPComputeBackendServiceWorkflow ingests GCP Compute backend services for a single project.
func GCPComputeBackendServiceWorkflow(ctx workflow.Context, params GCPComputeBackendServiceWorkflowParams) (*GCPComputeBackendServiceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeBackendServiceWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeBackendServicesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeBackendServicesActivity, IngestComputeBackendServicesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest backend services", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeBackendServiceWorkflow",
		"projectID", params.ProjectID,
		"backendServiceCount", result.BackendServiceCount,
	)

	return &GCPComputeBackendServiceWorkflowResult{
		ProjectID:           result.ProjectID,
		BackendServiceCount: result.BackendServiceCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
