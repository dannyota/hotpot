package agent

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// S1AgentWorkflowResult contains the result of the agent workflow.
type S1AgentWorkflowResult struct {
	AgentCount     int
	DurationMillis int64
}

// S1AgentWorkflow ingests SentinelOne agents.
func S1AgentWorkflow(ctx workflow.Context) (*S1AgentWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1AgentWorkflow")

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

	var result IngestS1AgentsResult
	err := workflow.ExecuteActivity(activityCtx, IngestS1AgentsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest agents", "error", err)
		return nil, err
	}

	logger.Info("Completed S1AgentWorkflow", "agentCount", result.AgentCount)

	return &S1AgentWorkflowResult{
		AgentCount:     result.AgentCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
