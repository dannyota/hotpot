package firewall

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOFirewallWorkflowResult contains the result of the Firewall workflow.
type DOFirewallWorkflowResult struct {
	FirewallCount  int
	DurationMillis int64
}

// DOFirewallWorkflow ingests DigitalOcean Firewalls.
func DOFirewallWorkflow(ctx workflow.Context) (*DOFirewallWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOFirewallWorkflow")

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

	var result IngestDOFirewallsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOFirewallsActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest firewalls", "error", err)
		return nil, err
	}

	logger.Info("Completed DOFirewallWorkflow", "firewallCount", result.FirewallCount)

	return &DOFirewallWorkflowResult{
		FirewallCount:  result.FirewallCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
