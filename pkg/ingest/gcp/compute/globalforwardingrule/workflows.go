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
// Creates its own session to manage client lifetime.
func GCPComputeGlobalForwardingRuleWorkflow(ctx workflow.Context, params GCPComputeGlobalForwardingRuleWorkflowParams) (*GCPComputeGlobalForwardingRuleWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeGlobalForwardingRuleWorkflow", "projectID", params.ProjectID)

	// Create session for client management
	sessionOpts := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 15 * time.Minute,
	}
	sess, err := workflow.CreateSession(ctx, sessionOpts)
	if err != nil {
		return nil, err
	}

	sessionInfo := workflow.GetSessionInfo(sess)
	sessionID := sessionInfo.SessionID

	// Ensure cleanup
	defer func() {
		workflow.ExecuteActivity(
			workflow.WithActivityOptions(sess, workflow.ActivityOptions{
				StartToCloseTimeout: time.Minute,
			}),
			CloseSessionClientActivity,
			CloseSessionClientParams{SessionID: sessionID},
		)
		workflow.CompleteSession(sess)
	}()

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
	sessCtx := workflow.WithActivityOptions(sess, activityOpts)

	// Execute ingest activity
	var result IngestComputeGlobalForwardingRulesResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeGlobalForwardingRulesActivity, IngestComputeGlobalForwardingRulesParams{
		SessionID: sessionID,
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
