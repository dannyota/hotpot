package asset

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// GCPCloudAssetAssetWorkflowParams contains parameters for the Cloud Asset asset workflow.
type GCPCloudAssetAssetWorkflowParams struct {
}

// GCPCloudAssetAssetWorkflowResult contains the result of the Cloud Asset asset workflow.
type GCPCloudAssetAssetWorkflowResult struct {
	AssetCount int
}

// GCPCloudAssetAssetWorkflow ingests Cloud Asset assets.
func GCPCloudAssetAssetWorkflow(ctx workflow.Context, params GCPCloudAssetAssetWorkflowParams) (*GCPCloudAssetAssetWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudAssetAssetWorkflow")

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

	var result IngestAssetsResult
	err := workflow.ExecuteActivity(activityCtx, IngestAssetsActivity, IngestAssetsParams{}).Get(ctx, &result)
	if err != nil {
		logger.Error("Failed to ingest Cloud Asset assets", "error", err)
		return nil, err
	}

	logger.Info("Completed GCPCloudAssetAssetWorkflow",
		"assetCount", result.AssetCount,
	)

	return &GCPCloudAssetAssetWorkflowResult{
		AssetCount: result.AssetCount,
	}, nil
}
