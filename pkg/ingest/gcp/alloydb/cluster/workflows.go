package cluster

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPAlloyDBClusterWorkflowParams contains parameters for the AlloyDB cluster workflow.
type GCPAlloyDBClusterWorkflowParams struct {
	ProjectID string
}

// GCPAlloyDBClusterWorkflowResult contains the result of the AlloyDB cluster workflow.
type GCPAlloyDBClusterWorkflowResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// GCPAlloyDBClusterWorkflow ingests AlloyDB clusters for a single project.
func GCPAlloyDBClusterWorkflow(ctx workflow.Context, params GCPAlloyDBClusterWorkflowParams) (*GCPAlloyDBClusterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAlloyDBClusterWorkflow", "projectID", params.ProjectID)

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

	var result IngestAlloyDBClustersResult
	err := workflow.ExecuteActivity(activityCtx, IngestAlloyDBClustersActivity, IngestAlloyDBClustersParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest AlloyDB clusters", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPAlloyDBClusterWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return &GCPAlloyDBClusterWorkflowResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
