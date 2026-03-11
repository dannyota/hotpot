package ranger_gateway

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// S1RangerGatewayWorkflowResult contains the result of the ranger gateway workflow.
type S1RangerGatewayWorkflowResult struct {
	GatewayCount   int
	DurationMillis int64
}

// S1RangerGatewayWorkflow ingests SentinelOne ranger gateways.
func S1RangerGatewayWorkflow(ctx workflow.Context) (*S1RangerGatewayWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1RangerGatewayWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestS1RangerGatewaysResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1RangerGatewaysActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest ranger gateways", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed S1RangerGatewayWorkflow", "gatewayCount", result.GatewayCount)

	return &S1RangerGatewayWorkflowResult{
		GatewayCount:   result.GatewayCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
