package address

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeAddressWorkflowParams contains parameters for the address workflow.
type GCPComputeAddressWorkflowParams struct {
	ProjectID string
}

// GCPComputeAddressWorkflowResult contains the result of the address workflow.
type GCPComputeAddressWorkflowResult struct {
	ProjectID      string
	AddressCount   int
	DurationMillis int64
}

// GCPComputeAddressWorkflow ingests GCP Compute regional addresses for a single project.
func GCPComputeAddressWorkflow(ctx workflow.Context, params GCPComputeAddressWorkflowParams) (*GCPComputeAddressWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeAddressWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeAddressesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeAddressesActivity, IngestComputeAddressesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest addresses", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeAddressWorkflow",
		"projectID", params.ProjectID,
		"addressCount", result.AddressCount,
	)

	return &GCPComputeAddressWorkflowResult{
		ProjectID:      result.ProjectID,
		AddressCount:   result.AddressCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
