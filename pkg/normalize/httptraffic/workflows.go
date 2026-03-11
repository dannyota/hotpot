package httptraffic

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// NormalizeHttptrafficWorkflowResult holds the workflow result.
type NormalizeHttptrafficWorkflowResult struct {
	TrafficResult   NormalizeTrafficResult
	UserAgentResult NormalizeUserAgentsResult
	ClientIPResult  NormalizeClientIPsResult
}

// NormalizeHttptrafficWorkflow normalizes bronze traffic data to silver.
func NormalizeHttptrafficWorkflow(ctx workflow.Context) (*NormalizeHttptrafficWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting NormalizeHttptrafficWorkflow")

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
	params := NormalizeTrafficParams{SinceMinutes: 30}

	// Compute Since once so all activities use the same cutoff time.
	sinceMinutes := params.SinceMinutes
	if sinceMinutes <= 0 {
		sinceMinutes = 30
	}
	params.Since = workflow.Now(ctx).Add(-time.Duration(sinceMinutes) * time.Minute)

	result := &NormalizeHttptrafficWorkflowResult{}

	// Step 1: Normalize traffic (must run first — UA + IP depend on same bronze window).
	if err := workflow.ExecuteActivity(activityCtx, NormalizeTrafficActivity, params).
		Get(ctx, &result.TrafficResult); err != nil {
		logger.Error("Failed to normalize traffic", "error", err)
		return nil, err
	}
	logger.Info("NormalizeTraffic done",
		"processed", result.TrafficResult.Processed,
		"mapped", result.TrafficResult.Mapped)

	// Step 2: Normalize UAs + IPs in parallel (independent — write different tables).
	var uaErr, ipErr error
	workflow.Go(ctx, func(gCtx workflow.Context) {
		uaErr = workflow.ExecuteActivity(
			workflow.WithActivityOptions(gCtx, activityOpts),
			NormalizeUserAgentsActivity, params,
		).Get(gCtx, &result.UserAgentResult)
	})
	workflow.Go(ctx, func(gCtx workflow.Context) {
		ipErr = workflow.ExecuteActivity(
			workflow.WithActivityOptions(gCtx, activityOpts),
			NormalizeClientIPsActivity, params,
		).Get(gCtx, &result.ClientIPResult)
	})

	// workflow.Go goroutines complete before workflow returns.
	if uaErr != nil {
		return nil, fmt.Errorf("normalize user agents: %w", uaErr)
	}
	if ipErr != nil {
		return nil, fmt.Errorf("normalize client IPs: %w", ipErr)
	}

	logger.Info("Completed NormalizeHttptrafficWorkflow",
		"trafficProcessed", result.TrafficResult.Processed,
		"uaProcessed", result.UserAgentResult.Processed,
		"ipProcessed", result.ClientIPResult.Processed)

	return result, nil
}
