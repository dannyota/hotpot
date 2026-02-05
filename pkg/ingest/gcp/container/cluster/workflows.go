package cluster

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPContainerClusterWorkflowParams contains parameters for the cluster workflow.
type GCPContainerClusterWorkflowParams struct {
	ProjectID string
}

// GCPContainerClusterWorkflowResult contains the result of the cluster workflow.
type GCPContainerClusterWorkflowResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// GCPContainerClusterWorkflow ingests GKE clusters for a single project.
// Creates its own session to manage client lifetime.
func GCPContainerClusterWorkflow(ctx workflow.Context, params GCPContainerClusterWorkflowParams) (*GCPContainerClusterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerClusterWorkflow", "projectID", params.ProjectID)

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
	var result IngestContainerClustersResult
	err = workflow.ExecuteActivity(sessCtx, IngestContainerClustersActivity, IngestContainerClustersParams{
		SessionID: sessionID,
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest clusters", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPContainerClusterWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return &GCPContainerClusterWorkflowResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
