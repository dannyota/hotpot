package globaladdress

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeGlobalAddressWorkflowParams contains parameters for the global address workflow.
type GCPComputeGlobalAddressWorkflowParams struct {
	ProjectID string
}

// GCPComputeGlobalAddressWorkflowResult contains the result of the global address workflow.
type GCPComputeGlobalAddressWorkflowResult struct {
	ProjectID          string
	GlobalAddressCount int
	DurationMillis     int64
}

// GCPComputeGlobalAddressWorkflow ingests GCP Compute global addresses for a single project.
func GCPComputeGlobalAddressWorkflow(ctx workflow.Context, params GCPComputeGlobalAddressWorkflowParams) (*GCPComputeGlobalAddressWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeGlobalAddressWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeGlobalAddressesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeGlobalAddressesActivity, IngestComputeGlobalAddressesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest global addresses", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeGlobalAddressWorkflow",
		"projectID", params.ProjectID,
		"globalAddressCount", result.GlobalAddressCount,
	)

	return &GCPComputeGlobalAddressWorkflowResult{
		ProjectID:          result.ProjectID,
		GlobalAddressCount: result.GlobalAddressCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}
