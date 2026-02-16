package appengine

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine/application"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine/appservice"
)

// GCPAppEngineWorkflowParams contains parameters for the App Engine workflow.
type GCPAppEngineWorkflowParams struct {
	ProjectID string
}

// GCPAppEngineWorkflowResult contains the result of the App Engine workflow.
type GCPAppEngineWorkflowResult struct {
	ProjectID        string
	ApplicationCount int
	ServiceCount     int
}

// GCPAppEngineWorkflow ingests all App Engine resources for a single project.
// Executes application workflow first, then services (services depend on application existing).
func GCPAppEngineWorkflow(ctx workflow.Context, params GCPAppEngineWorkflowParams) (*GCPAppEngineWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPAppEngineWorkflow", "projectID", params.ProjectID)

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

	result := &GCPAppEngineWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest application first (services reference application)
	var appResult application.GCPAppEngineApplicationWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, application.GCPAppEngineApplicationWorkflow,
		application.GCPAppEngineApplicationWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &appResult)
	if err != nil {
		logger.Error("Failed to ingest App Engine application", "error", err)
		return nil, err
	}
	result.ApplicationCount = appResult.ApplicationCount

	// Phase 2: Ingest services (depends on application being in DB)
	// Only proceed if an application exists
	if appResult.ApplicationCount > 0 {
		var svcResult appservice.GCPAppEngineServiceWorkflowResult
		err = workflow.ExecuteChildWorkflow(childCtx, appservice.GCPAppEngineServiceWorkflow,
			appservice.GCPAppEngineServiceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &svcResult)
		if err != nil {
			logger.Error("Failed to ingest App Engine services", "error", err)
			return nil, err
		}
		result.ServiceCount = svcResult.ServiceCount
	}

	logger.Info("Completed GCPAppEngineWorkflow",
		"projectID", params.ProjectID,
		"applicationCount", result.ApplicationCount,
		"serviceCount", result.ServiceCount,
	)

	return result, nil
}
