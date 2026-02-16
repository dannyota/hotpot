package resourcesearch

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPCloudAssetResourceSearchWorkflowParams contains parameters for the resource search workflow.
type GCPCloudAssetResourceSearchWorkflowParams struct {
}

// GCPCloudAssetResourceSearchWorkflowResult contains the result of the resource search workflow.
type GCPCloudAssetResourceSearchWorkflowResult struct {
	ResourceCount int
}

// GCPCloudAssetResourceSearchWorkflow ingests resource search results.
func GCPCloudAssetResourceSearchWorkflow(ctx workflow.Context, params GCPCloudAssetResourceSearchWorkflowParams) (*GCPCloudAssetResourceSearchWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudAssetResourceSearchWorkflow")

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

	var result IngestResourceSearchResult
	err := workflow.ExecuteActivity(activityCtx, IngestResourceSearchActivity, IngestResourceSearchParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest resource search results", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPCloudAssetResourceSearchWorkflow",
		"resourceCount", result.ResourceCount,
	)

	return &GCPCloudAssetResourceSearchWorkflowResult{
		ResourceCount: result.ResourceCount,
	}, nil
}
