package lb

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeLoadBalancerLBWorkflowParams contains parameters for the LB workflow.
type GreenNodeLoadBalancerLBWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeLoadBalancerLBWorkflowResult contains the result of the LB workflow.
type GreenNodeLoadBalancerLBWorkflowResult struct {
	LBCount        int
	DurationMillis int64
}

// GreenNodeLoadBalancerLBWorkflow ingests GreenNode load balancers.
func GreenNodeLoadBalancerLBWorkflow(ctx workflow.Context, params GreenNodeLoadBalancerLBWorkflowParams) (*GreenNodeLoadBalancerLBWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeLoadBalancerLBWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestLoadBalancerLBsResult
	err := workflow.ExecuteActivity(activityCtx, IngestLoadBalancerLBsActivity, IngestLoadBalancerLBsParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest load balancers", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeLoadBalancerLBWorkflow",
		"lbCount", result.LBCount,
	)

	return &GreenNodeLoadBalancerLBWorkflowResult{
		LBCount:        result.LBCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
