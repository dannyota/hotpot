package sql

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/sql/instance"
)

// GCPSQLWorkflowParams contains parameters for the SQL workflow.
type GCPSQLWorkflowParams struct {
	ProjectID string
}

// GCPSQLWorkflowResult contains the result of the SQL workflow.
type GCPSQLWorkflowResult struct {
	ProjectID     string
	InstanceCount int
}

// GCPSQLWorkflow ingests all GCP Cloud SQL resources for a single project.
// Orchestrates child workflows - each manages its own session and client lifecycle.
func GCPSQLWorkflow(ctx workflow.Context, params GCPSQLWorkflowParams) (*GCPSQLWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSQLWorkflow", "projectID", params.ProjectID)

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

	result := &GCPSQLWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Execute instance workflow
	var instanceResult instance.GCPSQLInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPSQLInstanceWorkflow,
		instance.GCPSQLInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest SQL instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	logger.Info("Completed GCPSQLWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
	)

	return result, nil
}
