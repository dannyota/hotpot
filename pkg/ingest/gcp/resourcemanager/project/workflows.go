package project

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerProjectWorkflowParams contains parameters for the project workflow.
type GCPResourceManagerProjectWorkflowParams struct {
	// Empty - discovers all accessible projects
}

// GCPResourceManagerProjectWorkflowResult contains the result of the project workflow.
type GCPResourceManagerProjectWorkflowResult struct {
	ProjectCount   int
	ProjectIDs     []string
	DurationMillis int64
}

// GCPResourceManagerProjectWorkflow discovers all GCP projects accessible by the service account.
// Creates its own session to manage client lifetime.
func GCPResourceManagerProjectWorkflow(ctx workflow.Context, params GCPResourceManagerProjectWorkflowParams) (*GCPResourceManagerProjectWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerProjectWorkflow")

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
	var result IngestProjectsResult
	err = workflow.ExecuteActivity(sessCtx, IngestProjectsActivity, IngestProjectsParams{
		SessionID: sessionID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to discover projects", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerProjectWorkflow",
		"projectCount", result.ProjectCount,
	)

	return &GCPResourceManagerProjectWorkflowResult{
		ProjectCount:   result.ProjectCount,
		ProjectIDs:     result.ProjectIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
