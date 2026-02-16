package project

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// DOProjectWorkflowResult contains the result of the Project workflow.
type DOProjectWorkflowResult struct {
	ProjectCount   int
	ResourceCount  int
	DurationMillis int64
}

// DOProjectWorkflow ingests DigitalOcean Projects and their Resources.
func DOProjectWorkflow(ctx workflow.Context) (*DOProjectWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting DOProjectWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Step 1: Ingest projects
	var projectsResult IngestDOProjectsResult
	err := workflow.ExecuteActivity(activityCtx, IngestDOProjectsActivity).Get(ctx, &projectsResult)
	if err != nil {
		logger.Error("Failed to ingest projects", "error", err)
		return nil, err
	}

	// Step 2: Ingest project resources using the project IDs from step 1
	var resourcesResult IngestDOProjectResourcesResult
	if len(projectsResult.ProjectIDs) > 0 {
		err = workflow.ExecuteActivity(activityCtx, IngestDOProjectResourcesActivity, IngestDOProjectResourcesInput{
			ProjectIDs: projectsResult.ProjectIDs,
		}).Get(ctx, &resourcesResult)
		if err != nil {
			logger.Error("Failed to ingest project resources", "error", err)
			return nil, err
		}
	}

	logger.Info("Completed DOProjectWorkflow",
		"projectCount", projectsResult.ProjectCount,
		"resourceCount", resourcesResult.ResourceCount,
	)

	return &DOProjectWorkflowResult{
		ProjectCount:   projectsResult.ProjectCount,
		ResourceCount:  resourcesResult.ResourceCount,
		DurationMillis: projectsResult.DurationMillis + resourcesResult.DurationMillis,
	}, nil
}
