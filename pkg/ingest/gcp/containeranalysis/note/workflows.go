package note

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPContainerAnalysisNoteWorkflowParams contains parameters for the note workflow.
type GCPContainerAnalysisNoteWorkflowParams struct {
	ProjectID string
}

// GCPContainerAnalysisNoteWorkflowResult contains the result of the note workflow.
type GCPContainerAnalysisNoteWorkflowResult struct {
	ProjectID      string
	NoteCount      int
	DurationMillis int64
}

// GCPContainerAnalysisNoteWorkflow ingests Grafeas notes for a single project.
func GCPContainerAnalysisNoteWorkflow(ctx workflow.Context, params GCPContainerAnalysisNoteWorkflowParams) (*GCPContainerAnalysisNoteWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerAnalysisNoteWorkflow", "projectID", params.ProjectID)

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

	var result IngestNotesResult
	err := workflow.ExecuteActivity(activityCtx, IngestNotesActivity, IngestNotesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest notes", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPContainerAnalysisNoteWorkflow",
		"projectID", params.ProjectID,
		"noteCount", result.NoteCount,
	)

	return &GCPContainerAnalysisNoteWorkflowResult{
		ProjectID:      result.ProjectID,
		NoteCount:      result.NoteCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
