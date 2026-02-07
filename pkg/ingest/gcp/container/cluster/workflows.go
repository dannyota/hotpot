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
func GCPContainerClusterWorkflow(ctx workflow.Context, params GCPContainerClusterWorkflowParams) (*GCPContainerClusterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerClusterWorkflow", "projectID", params.ProjectID)

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
	var result IngestContainerClustersResult
	err := workflow.ExecuteActivity(activityCtx, IngestContainerClustersActivity, IngestContainerClustersParams{
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
