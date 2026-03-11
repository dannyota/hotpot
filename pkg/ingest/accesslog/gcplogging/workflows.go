package gcplogging

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
	"danny.vn/hotpot/pkg/ingest/accesslog"
)

// GcpLoggingTrafficWorkflow ingests traffic counts from a single GCP Cloud Logging source.
func GcpLoggingTrafficWorkflow(ctx workflow.Context, params accesslog.ServiceWorkflowParams) (*accesslog.ServiceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GcpLoggingTrafficWorkflow", "sourceID", params.Name)

	// Use a longer timeout when backfill is configured.
	activityTimeout := 20 * time.Minute
	if params.BackfillDays > 0 {
		activityTimeout = 2 * time.Hour
	}
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: activityTimeout,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result accesslog.ServiceWorkflowResult
	err := workflow.ExecuteActivity(activityCtx, IngestTrafficCountsActivity, params).
		Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest traffic counts",
			"sourceID", params.Name, "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed GcpLoggingTrafficWorkflow",
		"sourceID", result.Name,
		"counts", result.Counts)

	return &result, nil
}
