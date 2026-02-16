package revision

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPRunRevisionWorkflowParams contains parameters for the Cloud Run revision workflow.
type GCPRunRevisionWorkflowParams struct {
	ProjectID string
}

// GCPRunRevisionWorkflowResult contains the result of the Cloud Run revision workflow.
type GCPRunRevisionWorkflowResult struct {
	ProjectID     string
	RevisionCount int
}

// GCPRunRevisionWorkflow ingests Cloud Run revisions for a single project.
func GCPRunRevisionWorkflow(ctx workflow.Context, params GCPRunRevisionWorkflowParams) (*GCPRunRevisionWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPRunRevisionWorkflow", "projectID", params.ProjectID)

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

	var result IngestRunRevisionsResult
	err := workflow.ExecuteActivity(activityCtx, IngestRunRevisionsActivity, IngestRunRevisionsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Cloud Run revisions", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPRunRevisionWorkflow",
		"projectID", params.ProjectID,
		"revisionCount", result.RevisionCount,
	)

	return &GCPRunRevisionWorkflowResult{
		ProjectID:     result.ProjectID,
		RevisionCount: result.RevisionCount,
	}, nil
}
