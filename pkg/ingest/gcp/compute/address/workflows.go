package address

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeAddressWorkflowParams contains parameters for the address workflow.
type GCPComputeAddressWorkflowParams struct {
	ProjectID string
}

// GCPComputeAddressWorkflowResult contains the result of the address workflow.
type GCPComputeAddressWorkflowResult struct {
	ProjectID      string
	AddressCount   int
	DurationMillis int64
}

// GCPComputeAddressWorkflow ingests GCP Compute regional addresses for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeAddressWorkflow(ctx workflow.Context, params GCPComputeAddressWorkflowParams) (*GCPComputeAddressWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeAddressWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeAddressesResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeAddressesActivity, IngestComputeAddressesParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest addresses", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeAddressWorkflow",
		"projectID", params.ProjectID,
		"addressCount", result.AddressCount,
	)

	return &GCPComputeAddressWorkflowResult{
		ProjectID:      result.ProjectID,
		AddressCount:   result.AddressCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
