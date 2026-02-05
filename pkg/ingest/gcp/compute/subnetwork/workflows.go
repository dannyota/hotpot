package subnetwork

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeSubnetworkWorkflowParams contains parameters for the subnetwork workflow.
type GCPComputeSubnetworkWorkflowParams struct {
	ProjectID string
}

// GCPComputeSubnetworkWorkflowResult contains the result of the subnetwork workflow.
type GCPComputeSubnetworkWorkflowResult struct {
	ProjectID       string
	SubnetworkCount int
	DurationMillis  int64
}

// GCPComputeSubnetworkWorkflow ingests GCP Compute subnetworks for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeSubnetworkWorkflow(ctx workflow.Context, params GCPComputeSubnetworkWorkflowParams) (*GCPComputeSubnetworkWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeSubnetworkWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeSubnetworksResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeSubnetworksActivity, IngestComputeSubnetworksParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest subnetworks", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeSubnetworkWorkflow",
		"projectID", params.ProjectID,
		"subnetworkCount", result.SubnetworkCount,
	)

	return &GCPComputeSubnetworkWorkflowResult{
		ProjectID:       result.ProjectID,
		SubnetworkCount: result.SubnetworkCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
