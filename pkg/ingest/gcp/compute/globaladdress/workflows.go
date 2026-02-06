package globaladdress

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeGlobalAddressWorkflowParams contains parameters for the global address workflow.
type GCPComputeGlobalAddressWorkflowParams struct {
	ProjectID string
}

// GCPComputeGlobalAddressWorkflowResult contains the result of the global address workflow.
type GCPComputeGlobalAddressWorkflowResult struct {
	ProjectID          string
	GlobalAddressCount int
	DurationMillis     int64
}

// GCPComputeGlobalAddressWorkflow ingests GCP Compute global addresses for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeGlobalAddressWorkflow(ctx workflow.Context, params GCPComputeGlobalAddressWorkflowParams) (*GCPComputeGlobalAddressWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeGlobalAddressWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeGlobalAddressesResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeGlobalAddressesActivity, IngestComputeGlobalAddressesParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest global addresses", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeGlobalAddressWorkflow",
		"projectID", params.ProjectID,
		"globalAddressCount", result.GlobalAddressCount,
	)

	return &GCPComputeGlobalAddressWorkflowResult{
		ProjectID:          result.ProjectID,
		GlobalAddressCount: result.GlobalAddressCount,
		DurationMillis:     result.DurationMillis,
	}, nil
}
