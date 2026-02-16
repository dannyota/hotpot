package cloudasset

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/asset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/iampolicysearch"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset/resourcesearch"
)

// GCPCloudAssetWorkflowParams contains parameters for the Cloud Asset workflow.
type GCPCloudAssetWorkflowParams struct {
}

// GCPCloudAssetWorkflowResult contains the result of the Cloud Asset workflow.
type GCPCloudAssetWorkflowResult struct {
	AssetCount    int
	PolicyCount   int
	ResourceCount int
}

// GCPCloudAssetWorkflow ingests all Cloud Asset Inventory resources.
// All 3 resources are independent and run in parallel.
func GCPCloudAssetWorkflow(ctx workflow.Context, params GCPCloudAssetWorkflowParams) (*GCPCloudAssetWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPCloudAssetWorkflow")

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

	result := &GCPCloudAssetWorkflowResult{}

	// Run all 3 resources in parallel
	assetFuture := workflow.ExecuteChildWorkflow(childCtx, asset.GCPCloudAssetAssetWorkflow,
		asset.GCPCloudAssetAssetWorkflowParams{})
	iamPolicyFuture := workflow.ExecuteChildWorkflow(childCtx, iampolicysearch.GCPCloudAssetIAMPolicySearchWorkflow,
		iampolicysearch.GCPCloudAssetIAMPolicySearchWorkflowParams{})
	resourceSearchFuture := workflow.ExecuteChildWorkflow(childCtx, resourcesearch.GCPCloudAssetResourceSearchWorkflow,
		resourcesearch.GCPCloudAssetResourceSearchWorkflowParams{})

	// Collect asset results
	var assetResult asset.GCPCloudAssetAssetWorkflowResult
	if err := assetFuture.Get(ctx, &assetResult); err != nil {
		logger.Error("Failed to ingest Cloud Asset assets", "error", err)
		return nil, err
	}
	result.AssetCount = assetResult.AssetCount

	// Collect IAM policy search results
	var iamPolicyResult iampolicysearch.GCPCloudAssetIAMPolicySearchWorkflowResult
	if err := iamPolicyFuture.Get(ctx, &iamPolicyResult); err != nil {
		logger.Error("Failed to ingest Cloud Asset IAM policy searches", "error", err)
		return nil, err
	}
	result.PolicyCount = iamPolicyResult.PolicyCount

	// Collect resource search results
	var resourceSearchResult resourcesearch.GCPCloudAssetResourceSearchWorkflowResult
	if err := resourceSearchFuture.Get(ctx, &resourceSearchResult); err != nil {
		logger.Error("Failed to ingest Cloud Asset resource searches", "error", err)
		return nil, err
	}
	result.ResourceCount = resourceSearchResult.ResourceCount

	logger.Info("Completed GCPCloudAssetWorkflow",
		"assetCount", result.AssetCount,
		"policyCount", result.PolicyCount,
		"resourceCount", result.ResourceCount,
	)

	return result, nil
}
