package targetinstance

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPComputeTargetInstanceWorkflowParams contains parameters for the target instance workflow.
type GCPComputeTargetInstanceWorkflowParams struct {
	ProjectID string
}

// GCPComputeTargetInstanceWorkflowResult contains the result of the target instance workflow.
type GCPComputeTargetInstanceWorkflowResult struct {
	ProjectID           string
	TargetInstanceCount int
	DurationMillis      int64
}

// GCPComputeTargetInstanceWorkflow ingests GCP Compute target instances for a single project.
// Creates its own session to manage client lifetime.
func GCPComputeTargetInstanceWorkflow(ctx workflow.Context, params GCPComputeTargetInstanceWorkflowParams) (*GCPComputeTargetInstanceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeTargetInstanceWorkflow", "projectID", params.ProjectID)

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
	var result IngestComputeTargetInstancesResult
	err = workflow.ExecuteActivity(sessCtx, IngestComputeTargetInstancesActivity, IngestComputeTargetInstancesParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest target instances", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPComputeTargetInstanceWorkflow",
		"projectID", params.ProjectID,
		"targetInstanceCount", result.TargetInstanceCount,
	)

	return &GCPComputeTargetInstanceWorkflowResult{
		ProjectID:           result.ProjectID,
		TargetInstanceCount: result.TargetInstanceCount,
		DurationMillis:      result.DurationMillis,
	}, nil
}
