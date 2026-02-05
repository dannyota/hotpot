package network

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeNetworkWorkflowParams contains parameters for the network workflow.
type GCPComputeNetworkWorkflowParams struct {
	ProjectID string
}

// GCPComputeNetworkWorkflowResult contains the result of the network workflow.
type GCPComputeNetworkWorkflowResult struct {
	ProjectID      string
	NetworkCount   int
	DurationMillis int64
}

// GCPComputeNetworkWorkflow ingests GCP Compute networks for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeNetworkWorkflow(ctx workflow.Context, params GCPComputeNetworkWorkflowParams) (*GCPComputeNetworkWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeNetworkWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeNetworksResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeNetworksActivity, IngestComputeNetworksParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest networks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeNetworkWorkflow",
		"projectID", params.ProjectID,
		"networkCount", result.NetworkCount,
	)

	return &GCPComputeNetworkWorkflowResult{
		ProjectID:      result.ProjectID,
		NetworkCount:   result.NetworkCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
