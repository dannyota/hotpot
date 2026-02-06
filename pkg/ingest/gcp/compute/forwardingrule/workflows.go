package forwardingrule

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeForwardingRuleWorkflowParams contains parameters for the forwarding rule workflow.
type GCPComputeForwardingRuleWorkflowParams struct {
	ProjectID string
}

// GCPComputeForwardingRuleWorkflowResult contains the result of the forwarding rule workflow.
type GCPComputeForwardingRuleWorkflowResult struct {
	ProjectID          string
	ForwardingRuleCount int
	DurationMillis     int64
}

// GCPComputeForwardingRuleWorkflow ingests GCP Compute regional forwarding rules for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeForwardingRuleWorkflow(ctx workflow.Context, params GCPComputeForwardingRuleWorkflowParams) (*GCPComputeForwardingRuleWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeForwardingRuleWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeForwardingRulesResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeForwardingRulesActivity, IngestComputeForwardingRulesParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest forwarding rules", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeForwardingRuleWorkflow",
		"projectID", params.ProjectID,
		"forwardingRuleCount", result.ForwardingRuleCount,
	)

	return &GCPComputeForwardingRuleWorkflowResult{
		ProjectID:          result.ProjectID,
		ForwardingRuleCount: result.ForwardingRuleCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}
