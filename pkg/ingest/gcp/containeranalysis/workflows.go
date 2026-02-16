package containeranalysis

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/note"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis/occurrence"
)

// GCPContainerAnalysisWorkflowParams contains parameters for the Container Analysis workflow.
type GCPContainerAnalysisWorkflowParams struct {
	ProjectID string
}

// GCPContainerAnalysisWorkflowResult contains the result of the Container Analysis workflow.
type GCPContainerAnalysisWorkflowResult struct {
	ProjectID       string
	NoteCount       int
	OccurrenceCount int
}

// GCPContainerAnalysisWorkflow ingests all Container Analysis resources for a single project.
// Notes and Occurrences are independent and run in parallel.
func GCPContainerAnalysisWorkflow(ctx workflow.Context, params GCPContainerAnalysisWorkflowParams) (*GCPContainerAnalysisWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerAnalysisWorkflow", "projectID", params.ProjectID)

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

	result := &GCPContainerAnalysisWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Run notes and occurrences in parallel (they are independent)
	noteFuture := workflow.ExecuteChildWorkflow(childCtx, note.GCPContainerAnalysisNoteWorkflow,
		note.GCPContainerAnalysisNoteWorkflowParams{ProjectID: params.ProjectID})

	occurrenceFuture := workflow.ExecuteChildWorkflow(childCtx, occurrence.GCPContainerAnalysisOccurrenceWorkflow,
		occurrence.GCPContainerAnalysisOccurrenceWorkflowParams{ProjectID: params.ProjectID})

	// Collect note results
	var noteResult note.GCPContainerAnalysisNoteWorkflowResult
	if err := noteFuture.Get(ctx, &noteResult); err != nil {
		logger.Error("Failed to ingest notes", "error", err)
		return nil, err
	}
	result.NoteCount = noteResult.NoteCount

	// Collect occurrence results
	var occurrenceResult occurrence.GCPContainerAnalysisOccurrenceWorkflowResult
	if err := occurrenceFuture.Get(ctx, &occurrenceResult); err != nil {
		logger.Error("Failed to ingest occurrences", "error", err)
		return nil, err
	}
	result.OccurrenceCount = occurrenceResult.OccurrenceCount

	logger.Info("Completed GCPContainerAnalysisWorkflow",
		"projectID", params.ProjectID,
		"noteCount", result.NoteCount,
		"occurrenceCount", result.OccurrenceCount,
	)

	return result, nil
}
