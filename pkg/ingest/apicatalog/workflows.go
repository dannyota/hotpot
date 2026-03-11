package apicatalog

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/temporalerr"
)

// ApiCatalogWorkflowParams holds parameters for the API catalog workflow.
type ApiCatalogWorkflowParams struct {
	FilePath    string
	CSVData     []byte
	LogSourceID string
	SourceFile  string
}

// ApiCatalogWorkflowResult holds the result of the API catalog workflow.
type ApiCatalogWorkflowResult struct {
	Created int
	Updated int
}

// ApiCatalogWorkflow imports API endpoint data from CSV.
func ApiCatalogWorkflow(ctx workflow.Context, params ApiCatalogWorkflowParams) (*ApiCatalogWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting ApiCatalogWorkflow")

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

	var result ImportCSVResult
	err := workflow.ExecuteActivity(activityCtx, ImportCSVActivity, ImportCSVParams{
		FilePath:    params.FilePath,
		CSVData:     params.CSVData,
		LogSourceID: params.LogSourceID,
		SourceFile:  params.SourceFile,
	}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to import CSV", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Completed ApiCatalogWorkflow",
		"created", result.Created,
		"updated", result.Updated)

	return &ApiCatalogWorkflowResult{
		Created: result.Created,
		Updated: result.Updated,
	}, nil
}
