package dataset

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBigQueryDatasetWorkflowParams contains parameters for the dataset workflow.
type GCPBigQueryDatasetWorkflowParams struct {
	ProjectID string
}

// GCPBigQueryDatasetWorkflowResult contains the result of the dataset workflow.
type GCPBigQueryDatasetWorkflowResult struct {
	ProjectID      string
	DatasetCount   int
	DatasetIDs     []string
	DurationMillis int64
}

// GCPBigQueryDatasetWorkflow ingests BigQuery datasets for a single project.
func GCPBigQueryDatasetWorkflow(ctx workflow.Context, params GCPBigQueryDatasetWorkflowParams) (*GCPBigQueryDatasetWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigQueryDatasetWorkflow", "projectID", params.ProjectID)

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

	var result IngestBigQueryDatasetsResult
	err := workflow.ExecuteActivity(activityCtx, IngestBigQueryDatasetsActivity, IngestBigQueryDatasetsParams{
		ProjectID: params.ProjectID,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest BigQuery datasets", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBigQueryDatasetWorkflow",
		"projectID", params.ProjectID,
		"datasetCount", result.DatasetCount,
	)

	return &GCPBigQueryDatasetWorkflowResult{
		ProjectID:      result.ProjectID,
		DatasetCount:   result.DatasetCount,
		DatasetIDs:     result.DatasetIDs,
		DurationMillis: result.DurationMillis,
	}, nil
}
