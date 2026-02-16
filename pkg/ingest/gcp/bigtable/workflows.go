package bigtable

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable/cluster"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable/instance"
)

// GCPBigtableWorkflowParams contains parameters for the Bigtable workflow.
type GCPBigtableWorkflowParams struct {
	ProjectID string
}

// GCPBigtableWorkflowResult contains the result of the Bigtable workflow.
type GCPBigtableWorkflowResult struct {
	ProjectID     string
	InstanceCount int
	ClusterCount  int
}

// GCPBigtableWorkflow ingests all Bigtable resources for a single project.
// Executes instance workflow first, then clusters (clusters reference instances).
func GCPBigtableWorkflow(ctx workflow.Context, params GCPBigtableWorkflowParams) (*GCPBigtableWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigtableWorkflow", "projectID", params.ProjectID)

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

	result := &GCPBigtableWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest instances first (clusters reference instances)
	var instanceResult instance.GCPBigtableInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPBigtableInstanceWorkflow,
		instance.GCPBigtableInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest Bigtable instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	// Phase 2: Ingest clusters (depends on instances being in DB)
	var clusterResult cluster.GCPBigtableClusterWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, cluster.GCPBigtableClusterWorkflow,
		cluster.GCPBigtableClusterWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &clusterResult)
	if err != nil {
		logger.Error("Failed to ingest Bigtable clusters", "error", err)
		return nil, err
	}
	result.ClusterCount = clusterResult.ClusterCount

	logger.Info("Completed GCPBigtableWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"clusterCount", result.ClusterCount,
	)

	return result, nil
}
