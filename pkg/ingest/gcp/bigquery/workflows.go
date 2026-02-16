package bigquery

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/dataset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery/table"
)

// GCPBigQueryWorkflowParams contains parameters for the BigQuery workflow.
type GCPBigQueryWorkflowParams struct {
	ProjectID string
}

// GCPBigQueryWorkflowResult contains the result of the BigQuery workflow.
type GCPBigQueryWorkflowResult struct {
	ProjectID    string
	DatasetCount int
	TableCount   int
}

// GCPBigQueryWorkflow ingests all BigQuery resources for a single project.
// Executes dataset workflow first, then tables (tables depend on datasets).
func GCPBigQueryWorkflow(ctx workflow.Context, params GCPBigQueryWorkflowParams) (*GCPBigQueryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPBigQueryWorkflow", "projectID", params.ProjectID)

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

	result := &GCPBigQueryWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Ingest datasets first (tables reference datasets)
	var datasetResult dataset.GCPBigQueryDatasetWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, dataset.GCPBigQueryDatasetWorkflow,
		dataset.GCPBigQueryDatasetWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &datasetResult)
	if err != nil {
		logger.Error("Failed to ingest BigQuery datasets", "error", err)
		return nil, err
	}
	result.DatasetCount = datasetResult.DatasetCount

	// Phase 2: Ingest tables (depends on datasets being in DB for dataset IDs)
	var tableResult table.GCPBigQueryTableWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, table.GCPBigQueryTableWorkflow,
		table.GCPBigQueryTableWorkflowParams{
			ProjectID:  params.ProjectID,
			DatasetIDs: datasetResult.DatasetIDs,
		}).Get(ctx, &tableResult)
	if err != nil {
		logger.Error("Failed to ingest BigQuery tables", "error", err)
		return nil, err
	}
	result.TableCount = tableResult.TableCount

	logger.Info("Completed GCPBigQueryWorkflow",
		"projectID", params.ProjectID,
		"datasetCount", result.DatasetCount,
		"tableCount", result.TableCount,
	)

	return result, nil
}
