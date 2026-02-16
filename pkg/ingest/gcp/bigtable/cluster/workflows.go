package cluster

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBigtableClusterWorkflowParams contains parameters for the cluster workflow.
type GCPBigtableClusterWorkflowParams struct {
	ProjectID string
}

// GCPBigtableClusterWorkflowResult contains the result of the cluster workflow.
type GCPBigtableClusterWorkflowResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// GCPBigtableClusterWorkflow ingests Bigtable clusters for all instances in a project.
func GCPBigtableClusterWorkflow(ctx workflow.Context, params GCPBigtableClusterWorkflowParams) (*GCPBigtableClusterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigtableClusterWorkflow", "projectID", params.ProjectID)

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

	var result IngestBigtableClustersResult
	err := workflow.ExecuteActivity(activityCtx, IngestBigtableClustersActivity, IngestBigtableClustersParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Bigtable clusters", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBigtableClusterWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return &GCPBigtableClusterWorkflowResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
