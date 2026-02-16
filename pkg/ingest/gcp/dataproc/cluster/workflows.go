package cluster

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPDataprocClusterWorkflowParams contains parameters for the Dataproc cluster workflow.
type GCPDataprocClusterWorkflowParams struct {
	ProjectID string
}

// GCPDataprocClusterWorkflowResult contains the result of the Dataproc cluster workflow.
type GCPDataprocClusterWorkflowResult struct {
	ProjectID      string
	ClusterCount   int
	DurationMillis int64
}

// GCPDataprocClusterWorkflow ingests Dataproc clusters for a single project.
func GCPDataprocClusterWorkflow(ctx workflow.Context, params GCPDataprocClusterWorkflowParams) (*GCPDataprocClusterWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPDataprocClusterWorkflow", "projectID", params.ProjectID)

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

	var result IngestDataprocClustersResult
	err := workflow.ExecuteActivity(activityCtx, IngestDataprocClustersActivity, IngestDataprocClustersParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Dataproc clusters", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPDataprocClusterWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return &GCPDataprocClusterWorkflowResult{
		ProjectID:      result.ProjectID,
		ClusterCount:   result.ClusterCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
