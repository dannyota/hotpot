package spanner

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/spanner/database"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/spanner/instance"
)

// GCPSpannerWorkflowParams contains parameters for the Spanner workflow.
type GCPSpannerWorkflowParams struct {
	ProjectID string
}

// GCPSpannerWorkflowResult contains the result of the Spanner workflow.
type GCPSpannerWorkflowResult struct {
	ProjectID     string
	InstanceCount int
	DatabaseCount int
}

// GCPSpannerWorkflow ingests all GCP Spanner resources for a single project.
// Executes instance workflow first, then databases (databases are listed per-instance).
func GCPSpannerWorkflow(ctx workflow.Context, params GCPSpannerWorkflowParams) (*GCPSpannerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPSpannerWorkflow", "projectID", params.ProjectID)

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPSpannerWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest instances first (databases need instance names)
	var instanceResult instance.GCPSpannerInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPSpannerInstanceWorkflow,
		instance.GCPSpannerInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest Spanner instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	// Phase 2: Ingest databases (depends on instances being in DB, uses instance names)
	if len(instanceResult.InstanceNames) > 0 {
		var databaseResult database.GCPSpannerDatabaseWorkflowResult
		err = workflow.ExecuteChildWorkflow(childCtx, database.GCPSpannerDatabaseWorkflow,
			database.GCPSpannerDatabaseWorkflowParams{
				ProjectID:     params.ProjectID,
				InstanceNames: instanceResult.InstanceNames,
			}).Get(ctx, &databaseResult)
		if err != nil {
			logger.Error("Failed to ingest Spanner databases", "error", err)
			return nil, err
		}
		result.DatabaseCount = databaseResult.DatabaseCount
	}

	logger.Info("Completed GCPSpannerWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"databaseCount", result.DatabaseCount,
	)

	return result, nil
}
