package loadbalancer

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOLoadBalancerWorkflowResult contains the result of the Load Balancer workflow.
type DOLoadBalancerWorkflowResult struct {
	LoadBalancerCount int
	DurationMillis    int64
}

// DOLoadBalancerWorkflow ingests DigitalOcean Load Balancers.
func DOLoadBalancerWorkflow(ctx workflow.Context) (*DOLoadBalancerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOLoadBalancerWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestDOLoadBalancersResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOLoadBalancersActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest load balancers", "error", err)
		return nil, err
	}

	logger.Info("Completed DOLoadBalancerWorkflow", "loadBalancerCount", result.LoadBalancerCount)

	return &DOLoadBalancerWorkflowResult{
		LoadBalancerCount: result.LoadBalancerCount,
		DurationMillis:    result.DurationMillis,
	}, nil
}
