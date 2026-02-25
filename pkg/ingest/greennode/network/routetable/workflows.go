package routetable

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeNetworkRouteTableWorkflowParams contains parameters for the route table workflow.
type GreenNodeNetworkRouteTableWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkRouteTableWorkflowResult contains the result of the route table workflow.
type GreenNodeNetworkRouteTableWorkflowResult struct {
	RouteTableCount int
	DurationMillis  int64
}

// GreenNodeNetworkRouteTableWorkflow ingests GreenNode route tables.
func GreenNodeNetworkRouteTableWorkflow(ctx workflow.Context, params GreenNodeNetworkRouteTableWorkflowParams) (*GreenNodeNetworkRouteTableWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkRouteTableWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestNetworkRouteTablesResult
	err := workflow.ExecuteActivity(activityCtx, IngestNetworkRouteTablesActivity, IngestNetworkRouteTablesParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest route tables", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeNetworkRouteTableWorkflow",
		"routeTableCount", result.RouteTableCount,
	)

	return &GreenNodeNetworkRouteTableWorkflowResult{
		RouteTableCount: result.RouteTableCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
