package glbresource

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GreenNodeGLBGlobalLoadBalancerWorkflowParams contains parameters for the GLB workflow.
type GreenNodeGLBGlobalLoadBalancerWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeGLBGlobalLoadBalancerWorkflowResult contains the result of the GLB workflow.
type GreenNodeGLBGlobalLoadBalancerWorkflowResult struct {
	GLBCount       int
	DurationMillis int64
}

// GreenNodeGLBGlobalLoadBalancerWorkflow ingests GreenNode global load balancers.
func GreenNodeGLBGlobalLoadBalancerWorkflow(ctx workflow.Context, params GreenNodeGLBGlobalLoadBalancerWorkflowParams) (*GreenNodeGLBGlobalLoadBalancerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeGLBGlobalLoadBalancerWorkflow", "projectID", params.ProjectID, "region", params.Region)

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

	var result IngestGLBGlobalLoadBalancersResult
	err := workflow.ExecuteActivity(activityCtx, IngestGLBGlobalLoadBalancersActivity, IngestGLBGlobalLoadBalancersParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest GLBs", "error", err)
		return nil, err
	}

	logger.Info("Completed GreenNodeGLBGlobalLoadBalancerWorkflow",
		"glbCount", result.GLBCount,
	)

	return &GreenNodeGLBGlobalLoadBalancerWorkflowResult{
		GLBCount:       result.GLBCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
