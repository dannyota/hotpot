package container

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/container/cluster"
)

// GCPContainerWorkflowParams contains parameters for the container workflow.
type GCPContainerWorkflowParams struct {
	ProjectID string
}

// GCPContainerWorkflowResult contains the result of the container workflow.
type GCPContainerWorkflowResult struct {
	ProjectID    string
	ClusterCount int
}

// GCPContainerWorkflow ingests all GKE Container resources for a single project.
// Orchestrates child workflows - each manages its own session and client lifecycle.
func GCPContainerWorkflow(ctx workflow.Context, params GCPContainerWorkflowParams) (*GCPContainerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerWorkflow", "projectID", params.ProjectID)

	// Child workflow options
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

	result := &GCPContainerWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute cluster workflow
	var clusterResult cluster.GCPContainerClusterWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, cluster.GCPContainerClusterWorkflow,
		cluster.GCPContainerClusterWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &clusterResult)
	if err != nil {
		logger.Error("Failed to ingest clusters", "error", err)
		return nil, err
	}
	result.ClusterCount = clusterResult.ClusterCount

	logger.Info("Completed GCPContainerWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return result, nil
}
