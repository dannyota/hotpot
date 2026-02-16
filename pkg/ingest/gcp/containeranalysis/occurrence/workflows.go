package occurrence

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPContainerAnalysisOccurrenceWorkflowParams contains parameters for the occurrence workflow.
type GCPContainerAnalysisOccurrenceWorkflowParams struct {
	ProjectID string
}

// GCPContainerAnalysisOccurrenceWorkflowResult contains the result of the occurrence workflow.
type GCPContainerAnalysisOccurrenceWorkflowResult struct {
	ProjectID       string
	OccurrenceCount int
	DurationMillis  int64
}

// GCPContainerAnalysisOccurrenceWorkflow ingests Grafeas occurrences for a single project.
func GCPContainerAnalysisOccurrenceWorkflow(ctx workflow.Context, params GCPContainerAnalysisOccurrenceWorkflowParams) (*GCPContainerAnalysisOccurrenceWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPContainerAnalysisOccurrenceWorkflow", "projectID", params.ProjectID)

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result IngestOccurrencesResult
	err := workflow.ExecuteActivity(activityCtx, IngestOccurrencesActivity, IngestOccurrencesParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest occurrences", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPContainerAnalysisOccurrenceWorkflow",
		"projectID", params.ProjectID,
		"occurrenceCount", result.OccurrenceCount,
	)

	return &GCPContainerAnalysisOccurrenceWorkflowResult{
		ProjectID:       result.ProjectID,
		OccurrenceCount: result.OccurrenceCount,
		DurationMillis:  result.DurationMillis,
	}, nil
}
