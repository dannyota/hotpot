package interconnect

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeInterconnectWorkflowParams contains parameters for the interconnect workflow.
type GCPComputeInterconnectWorkflowParams struct {
	ProjectID string
}

// GCPComputeInterconnectWorkflowResult contains the result of the interconnect workflow.
type GCPComputeInterconnectWorkflowResult struct {
	ProjectID         string
	InterconnectCount int
	DurationMillis    int64
}

// GCPComputeInterconnectWorkflow ingests GCP Compute interconnects for a single project.
func GCPComputeInterconnectWorkflow(ctx workflow.Context, params GCPComputeInterconnectWorkflowParams) (*GCPComputeInterconnectWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeInterconnectWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeInterconnectsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeInterconnectsActivity, IngestComputeInterconnectsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest interconnects", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeInterconnectWorkflow",
		"projectID", params.ProjectID,
		"interconnectCount", result.InterconnectCount,
	)

	return &GCPComputeInterconnectWorkflowResult{
		ProjectID:         result.ProjectID,
		InterconnectCount: result.InterconnectCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
