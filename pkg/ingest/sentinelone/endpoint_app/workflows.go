package endpoint_app

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

const batchSize = 50

// S1EndpointAppWorkflowResult contains the result of the endpoint app workflow.
type S1EndpointAppWorkflowResult struct {
	AppCount       int
	DurationMillis int64
}

// S1EndpointAppWorkflow ingests SentinelOne endpoint applications in sequential batches.
func S1EndpointAppWorkflow(ctx workflow.Context) (*S1EndpointAppWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1EndpointAppWorkflow")

	startTime := workflow.Now(ctx)

	// Step 1: List agent IDs from database
	listCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	var listResult ListAgentIDsResult
	if err := workflow.ExecuteActivity(listCtx, ListAgentIDsActivity).Get(ctx, &listResult); err != nil {
		logger.Error("Failed to list agent IDs", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Listed agent IDs", "agentCount", len(listResult.AgentIDs))

	// Step 2: Process agents in sequential batches.
	// Each batch activity processes up to batchSize agents, with the rate
	// limiter pacing API calls. This avoids fan-out complexity and keeps
	// workflow history small.
	fetchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    5 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	})

	totalApps := 0
	for i := 0; i < len(listResult.AgentIDs); i += batchSize {
		end := i + batchSize
		if end > len(listResult.AgentIDs) {
			end = len(listResult.AgentIDs)
		}
		batch := listResult.AgentIDs[i:end]

		var result FetchAndSaveBatchResult
		err := workflow.ExecuteActivity(fetchCtx, FetchAndSaveBatchActivity, FetchAndSaveBatchInput{
			AgentIDs:    batch,
			CollectedAt: listResult.CollectedAt,
		}).Get(ctx, &result)
		if err != nil {
			logger.Error("Failed to process batch", "batchStart", i, "error", err)
			return nil, temporalerr.PropagateNonRetryable(err)
		}

		totalApps += result.AppCount
		logger.Info("s1 endpoint apps: batch complete",
			"agentsDone", end,
			"totalAgents", len(listResult.AgentIDs),
			"totalApps", totalApps,
		)
	}

	// Step 3: Delete orphan endpoint apps
	deleteCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	if err := workflow.ExecuteActivity(deleteCtx, DeleteOrphanEndpointAppsActivity).Get(ctx, nil); err != nil {
		logger.Warn("Failed to delete orphan endpoint apps", "error", err)
	}

	durationMillis := workflow.Now(ctx).Sub(startTime).Milliseconds()
	logger.Info("Completed S1EndpointAppWorkflow", "appCount", totalApps, "durationMillis", durationMillis)

	return &S1EndpointAppWorkflowResult{
		AppCount:       totalApps,
		DurationMillis: durationMillis,
	}, nil
}
