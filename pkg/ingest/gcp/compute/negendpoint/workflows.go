package negendpoint

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type GCPComputeNegEndpointWorkflowParams struct {
	ProjectID string
}

type GCPComputeNegEndpointWorkflowResult struct {
	ProjectID        string
	NegEndpointCount int
}

func GCPComputeNegEndpointWorkflow(ctx workflow.Context, params GCPComputeNegEndpointWorkflowParams) (*GCPComputeNegEndpointWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeNegEndpointWorkflow", "projectID", params.ProjectID)

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

	var result IngestComputeNegEndpointsResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeNegEndpointsActivity,
		IngestComputeNegEndpointsParams{ProjectID: params.ProjectID}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest NEG endpoints", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeNegEndpointWorkflow",
		"projectID", params.ProjectID,
		"negEndpointCount", result.NegEndpointCount,
	)

	return &GCPComputeNegEndpointWorkflowResult{
		ProjectID:        result.ProjectID,
		NegEndpointCount: result.NegEndpointCount,
	}, nil
}
