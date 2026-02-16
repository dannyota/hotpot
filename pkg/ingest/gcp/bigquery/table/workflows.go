package table

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPBigQueryTableWorkflowParams contains parameters for the table workflow.
type GCPBigQueryTableWorkflowParams struct {
	ProjectID  string
	DatasetIDs []string
}

// GCPBigQueryTableWorkflowResult contains the result of the table workflow.
type GCPBigQueryTableWorkflowResult struct {
	ProjectID      string
	TableCount     int
	DurationMillis int64
}

// GCPBigQueryTableWorkflow ingests BigQuery tables for a single project.
func GCPBigQueryTableWorkflow(ctx workflow.Context, params GCPBigQueryTableWorkflowParams) (*GCPBigQueryTableWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigQueryTableWorkflow",
		"projectID", params.ProjectID,
		"datasetCount", len(params.DatasetIDs),
	)

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

	var result IngestBigQueryTablesResult
	err := workflow.ExecuteActivity(activityCtx, IngestBigQueryTablesActivity, IngestBigQueryTablesParams{
		ProjectID:  params.ProjectID,
		DatasetIDs: params.DatasetIDs,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest BigQuery tables", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPBigQueryTableWorkflow",
		"projectID", params.ProjectID,
		"tableCount", result.TableCount,
	)

	return &GCPBigQueryTableWorkflowResult{
		ProjectID:      result.ProjectID,
		TableCount:     result.TableCount,
		DurationMillis: result.DurationMillis,
	}, nil
}
