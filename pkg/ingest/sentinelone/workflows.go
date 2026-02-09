package sentinelone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"hotpot/pkg/ingest/sentinelone/account"
	"hotpot/pkg/ingest/sentinelone/agent"
	"hotpot/pkg/ingest/sentinelone/threat"
)

// S1InventoryWorkflowResult contains the result of SentinelOne inventory collection.
type S1InventoryWorkflowResult struct {
	AccountCount int
	AgentCount   int
	ThreatCount  int
}

// S1InventoryWorkflow orchestrates SentinelOne inventory collection.
func S1InventoryWorkflow(ctx workflow.Context) (*S1InventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1InventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &S1InventoryWorkflowResult{}

	// Execute account workflow
	var accountResult account.S1AccountWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, account.S1AccountWorkflow).Get(ctx, &accountResult)
	if err != nil {
		logger.Error("Failed to execute S1AccountWorkflow", "error", err)
	} else {
		result.AccountCount = accountResult.AccountCount
	}

	// Execute agent workflow
	var agentResult agent.S1AgentWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, agent.S1AgentWorkflow).Get(ctx, &agentResult)
	if err != nil {
		logger.Error("Failed to execute S1AgentWorkflow", "error", err)
	} else {
		result.AgentCount = agentResult.AgentCount
	}

	// Execute threat workflow
	var threatResult threat.S1ThreatWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, threat.S1ThreatWorkflow).Get(ctx, &threatResult)
	if err != nil {
		logger.Error("Failed to execute S1ThreatWorkflow", "error", err)
	} else {
		result.ThreatCount = threatResult.ThreatCount
	}

	logger.Info("Completed S1InventoryWorkflow",
		"accounts", result.AccountCount,
		"agents", result.AgentCount,
		"threats", result.ThreatCount,
	)

	return result, nil
}
