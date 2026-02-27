package endpoint_app

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
)

// S1EndpointAppWorkflowResult contains the result of the endpoint app workflow.
type S1EndpointAppWorkflowResult struct {
	AppCount       int
	DurationMillis int64
}

// S1EndpointAppWorkflow ingests SentinelOne endpoint applications.
func S1EndpointAppWorkflow(ctx workflow.Context) (*S1EndpointAppWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1EndpointAppWorkflow")

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

	var result IngestS1EndpointAppsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1EndpointAppsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest endpoint apps", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed S1EndpointAppWorkflow", "appCount", result.AppCount)

	return &S1EndpointAppWorkflowResult{
		AppCount:       result.AppCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
