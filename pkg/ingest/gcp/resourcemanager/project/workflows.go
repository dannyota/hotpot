package project

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPResourceManagerProjectWorkflowParams contains parameters for the project workflow.
type GCPResourceManagerProjectWorkflowParams struct {
	// Empty - discovers all accessible projects
}

// GCPResourceManagerProjectWorkflowResult contains the result of the project workflow.
type GCPResourceManagerProjectWorkflowResult struct {
	ProjectCount   int
	ProjectIDs     []string
	DurationMillis int64
}

// GCPResourceManagerProjectWorkflow discovers all GCP projects accessible by the service account.
func GCPResourceManagerProjectWorkflow(ctx workflow.Context, params GCPResourceManagerProjectWorkflowParams) (*GCPResourceManagerProjectWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerProjectWorkflow")

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
	var result IngestProjectsResult
	err := workflow.ExecuteActivity(activityCtx, IngestProjectsActivity, IngestProjectsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to discover projects", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPResourceManagerProjectWorkflow",
		"projectCount", result.ProjectCount,
	)

	return &GCPResourceManagerProjectWorkflowResult{
		ProjectCount:   result.ProjectCount,
		ProjectIDs:     result.ProjectIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
