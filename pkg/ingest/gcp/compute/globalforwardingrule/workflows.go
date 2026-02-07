package globalforwardingrule

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeGlobalForwardingRuleWorkflowParams contains parameters for the global forwarding rule workflow.
type GCPComputeGlobalForwardingRuleWorkflowParams struct {
	ProjectID string
}

// GCPComputeGlobalForwardingRuleWorkflowResult contains the result of the global forwarding rule workflow.
type GCPComputeGlobalForwardingRuleWorkflowResult struct {
	ProjectID                string
	GlobalForwardingRuleCount int
	DurationMillis           int64
}

// GCPComputeGlobalForwardingRuleWorkflow ingests GCP Compute global forwarding rules for a single project.
func GCPComputeGlobalForwardingRuleWorkflow(ctx workflow.Context, params GCPComputeGlobalForwardingRuleWorkflowParams) (*GCPComputeGlobalForwardingRuleWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeGlobalForwardingRuleWorkflow", "projectID", params.ProjectID)

	// Activity options
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

	// Execute ingest activity
	var result IngestComputeGlobalForwardingRulesResult
	err := workflow.ExecuteActivity(activityCtx, IngestComputeGlobalForwardingRulesActivity, IngestComputeGlobalForwardingRulesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest global forwarding rules", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeGlobalForwardingRuleWorkflow",
		"projectID", params.ProjectID,
		"globalForwardingRuleCount", result.GlobalForwardingRuleCount,
	)

	return &GCPComputeGlobalForwardingRuleWorkflowResult{
		ProjectID:                result.ProjectID,
		GlobalForwardingRuleCount: result.GlobalForwardingRuleCount,
		DurationMillis:           result.DurationMillis,
	}, nil
}
