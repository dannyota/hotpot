package portal

import (
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/quota"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/region"
)

// GreenNodePortalWorkflowResult contains the result of the portal workflow.
type GreenNodePortalWorkflowResult struct {
	RegionCount int
	QuotaCount  int
}

// GreenNodePortalWorkflow orchestrates GreenNode portal ingestion.
func GreenNodePortalWorkflow(ctx workflow.Context) (*GreenNodePortalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalWorkflow")

	result := &GreenNodePortalWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Regions
	var regionResult region.GreenNodePortalRegionWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, region.GreenNodePortalRegionWorkflow).Get(ctx, &regionResult)
	if err != nil {
		logger.Error("Failed to ingest regions", "error", err)
	} else {
		result.RegionCount = regionResult.RegionCount
	}

	// Quotas
	var quotaResult quota.GreenNodePortalQuotaWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, quota.GreenNodePortalQuotaWorkflow).Get(ctx, &quotaResult)
	if err != nil {
		logger.Error("Failed to ingest quotas", "error", err)
	} else {
		result.QuotaCount = quotaResult.QuotaCount
	}

	logger.Info("Completed GreenNodePortalWorkflow",
		"regionCount", result.RegionCount,
		"quotaCount", result.QuotaCount,
	)

	return result, nil
}
