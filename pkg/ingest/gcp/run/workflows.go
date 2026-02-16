package run

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/revision"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run/service"
)

// GCPRunWorkflowParams contains parameters for the Cloud Run workflow.
type GCPRunWorkflowParams struct {
	ProjectID string
}

// GCPRunWorkflowResult contains the result of the Cloud Run workflow.
type GCPRunWorkflowResult struct {
	ProjectID     string
	ServiceCount  int
	RevisionCount int
}

// GCPRunWorkflow ingests all Cloud Run resources.
// Executes service workflow first, then revisions (revisions depend on services).
func GCPRunWorkflow(ctx workflow.Context, params GCPRunWorkflowParams) (*GCPRunWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPRunWorkflow", "projectID", params.ProjectID)

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

	result := &GCPRunWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest services first (revisions reference services)
	var serviceResult service.GCPRunServiceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, service.GCPRunServiceWorkflow,
		service.GCPRunServiceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &serviceResult)
	if err != nil {
		logger.Error("Failed to ingest Cloud Run services", "error", err)
		return nil, err
	}
	result.ServiceCount = serviceResult.ServiceCount

	// Phase 2: Ingest revisions (depends on services being in DB)
	var revisionResult revision.GCPRunRevisionWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, revision.GCPRunRevisionWorkflow,
		revision.GCPRunRevisionWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &revisionResult)
	if err != nil {
		logger.Error("Failed to ingest Cloud Run revisions", "error", err)
		return nil, err
	}
	result.RevisionCount = revisionResult.RevisionCount

	logger.Info("Completed GCPRunWorkflow",
		"projectID", params.ProjectID,
		"serviceCount", result.ServiceCount,
		"revisionCount", result.RevisionCount,
	)

	return result, nil
}
