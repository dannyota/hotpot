package endpoint

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkEndpointWorkflowParams contains parameters for the endpoint workflow.
type GreenNodeNetworkEndpointWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkEndpointWorkflowResult contains the result of the endpoint workflow.
type GreenNodeNetworkEndpointWorkflowResult struct {
	EndpointCount  int
	DurationMillis int64
}

// GreenNodeNetworkEndpointWorkflow ingests GreenNode endpoints.
func GreenNodeNetworkEndpointWorkflow(ctx workflow.Context, params GreenNodeNetworkEndpointWorkflowParams) (*GreenNodeNetworkEndpointWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkEndpointWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkEndpointsResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkEndpointsActivity, IngestNetworkEndpointsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest endpoints", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkEndpointWorkflow", "endpointCount", result.EndpointCount)

	return &GreenNodeNetworkEndpointWorkflowResult{
		EndpointCount:  result.EndpointCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
