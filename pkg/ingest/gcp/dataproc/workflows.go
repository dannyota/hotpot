package dataproc

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/dataproc/cluster"
)

// GCPDataprocWorkflowParams contains parameters for the Dataproc workflow.
type GCPDataprocWorkflowParams struct {
	ProjectID string
}

// GCPDataprocWorkflowResult contains the result of the Dataproc workflow.
type GCPDataprocWorkflowResult struct {
	ProjectID    string
	ClusterCount int
}

// GCPDataprocWorkflow ingests all GCP Dataproc resources for a single project.
func GCPDataprocWorkflow(ctx workflow.Context, params GCPDataprocWorkflowParams) (*GCPDataprocWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPDataprocWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPDataprocWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var clusterResult cluster.GCPDataprocClusterWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, cluster.GCPDataprocClusterWorkflow,
		cluster.GCPDataprocClusterWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &clusterResult)
	if err != nil {
		logger.Error("Failed to ingest Dataproc clusters", "error", err)
		return nil, err
	}
	result.ClusterCount = clusterResult.ClusterCount

	logger.Info("Completed GCPDataprocWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return result, nil
}
