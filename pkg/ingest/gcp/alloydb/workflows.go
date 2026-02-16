package alloydb

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/alloydb/cluster"
)

// GCPAlloyDBWorkflowParams contains parameters for the AlloyDB workflow.
type GCPAlloyDBWorkflowParams struct {
	ProjectID string
}

// GCPAlloyDBWorkflowResult contains the result of the AlloyDB workflow.
type GCPAlloyDBWorkflowResult struct {
	ProjectID    string
	ClusterCount int
}

// GCPAlloyDBWorkflow ingests all GCP AlloyDB resources for a single project.
func GCPAlloyDBWorkflow(ctx workflow.Context, params GCPAlloyDBWorkflowParams) (*GCPAlloyDBWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAlloyDBWorkflow", "projectID", params.ProjectID)

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

	result := &GCPAlloyDBWorkflowResult{
		ProjectID: params.ProjectID,
	}

	var clusterResult cluster.GCPAlloyDBClusterWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, cluster.GCPAlloyDBClusterWorkflow,
		cluster.GCPAlloyDBClusterWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &clusterResult)
	if err != nil {
		logger.Error("Failed to ingest AlloyDB clusters", "error", err)
		return nil, err
	}
	result.ClusterCount = clusterResult.ClusterCount

	logger.Info("Completed GCPAlloyDBWorkflow",
		"projectID", params.ProjectID,
		"clusterCount", result.ClusterCount,
	)

	return result, nil
}
