package resourcemanager

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"hotpot/pkg/ingest/gcp/resourcemanager/project"
)

// GCPResourceManagerWorkflowParams contains parameters for the resource manager workflow.
type GCPResourceManagerWorkflowParams struct {
	// Empty - discovers all accessible resources
}

// GCPResourceManagerWorkflowResult contains the result of the resource manager workflow.
type GCPResourceManagerWorkflowResult struct {
	ProjectCount   int
	ProjectIDs     []string
	DurationMillis int64
	// FolderCount       int      // future
	// OrganizationCount int      // future
}

// GCPResourceManagerWorkflow ingests all GCP Resource Manager resources.
// Orchestrates child workflows - each manages its own session and client lifecycle.
func GCPResourceManagerWorkflow(ctx workflow.Context, params GCPResourceManagerWorkflowParams) (*GCPResourceManagerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerWorkflow")

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

	result := &GCPResourceManagerWorkflowResult{}
	startTime := workflow.Now(ctx)

	// Execute project workflow
	var projectResult project.GCPResourceManagerProjectWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, project.GCPResourceManagerProjectWorkflow,
		project.GCPResourceManagerProjectWorkflowParams{}).Get(ctx, &projectResult)
	if err != nil {
		logger.Error("Failed to ingest projects", "error", err)
		return nil, err
	}
	result.ProjectCount = projectResult.ProjectCount
	result.ProjectIDs = projectResult.ProjectIDs

	// Execute folder workflow (future)
	// var folderResult folder.GCPResourceManagerFolderWorkflowResult
	// err = workflow.ExecuteChildWorkflow(childCtx, folder.GCPResourceManagerFolderWorkflow, ...).Get(ctx, &folderResult)

	// Execute organization workflow (future)
	// var orgResult organization.GCPResourceManagerOrganizationWorkflowResult
	// err = workflow.ExecuteChildWorkflow(childCtx, organization.GCPResourceManagerOrganizationWorkflow, ...).Get(ctx, &orgResult)

	result.DurationMillis = workflow.Now(ctx).Sub(startTime).Milliseconds()

	logger.Info("Completed GCPResourceManagerWorkflow",
		"projectCount", result.ProjectCount,
	)

	return result, nil
}
