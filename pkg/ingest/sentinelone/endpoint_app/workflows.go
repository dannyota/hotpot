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

// S1EndpointAppWorkflow ingests SentinelOne endpoint applications using per-agent fan-out.
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

	// Step 2: Fan-out — fetch and save apps per agent
	fetchCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	totalApps := 0
	agentsDone := 0
	totalAgents := len(listResult.AgentIDs)
	sem := workflow.NewSemaphore(ctx, 5)
	mu := workflow.NewMutex(ctx)
	wg := workflow.NewWaitGroup(ctx)
	var firstErr error

	for _, agentID := range listResult.AgentIDs {
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}

		wg.Add(1)
		agentID := agentID // capture loop variable
		workflow.Go(ctx, func(gCtx workflow.Context) {
			defer wg.Done()
			defer sem.Release(1)

			var result FetchAndSaveAgentAppsResult
			err := workflow.ExecuteActivity(fetchCtx, FetchAndSaveAgentAppsActivity, FetchAndSaveAgentAppsInput{
				AgentID:     agentID,
				CollectedAt: listResult.CollectedAt,
			}).Get(gCtx, &result)

			_ = mu.Lock(gCtx)
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			totalApps += result.AppCount
			agentsDone++
			if agentsDone%50 == 0 || agentsDone == totalAgents {
				logger.Info("s1 endpoint apps: progress", "agentsDone", agentsDone, "totalAgents", totalAgents, "totalApps", totalApps)
			}
		})
	}

	wg.Wait(ctx)

	if firstErr != nil {
		logger.Error("Failed to fetch agent apps", "error", firstErr)
		return nil, temporalerr.PropagateNonRetryable(firstErr)
	}

	logger.Info("Fetched all agent apps", "totalApps", totalApps, "agentCount", len(listResult.AgentIDs))

	// Step 3: Delete stale endpoint apps
	deleteCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	})

	if err := workflow.ExecuteActivity(deleteCtx, DeleteStaleEndpointAppsActivity, DeleteStaleEndpointAppsInput{
		CollectedAt: listResult.CollectedAt,
	}).Get(ctx, nil); err != nil {
		logger.Warn("Failed to delete stale endpoint apps", "error", err)
	}

	durationMillis := workflow.Now(ctx).Sub(startTime).Milliseconds()
	logger.Info("Completed S1EndpointAppWorkflow", "appCount", totalApps, "durationMillis", durationMillis)

	return &S1EndpointAppWorkflowResult{
		AppCount:       totalApps,
		DurationMillis: durationMillis,
	}, nil
}
