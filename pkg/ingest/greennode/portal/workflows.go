package portal

import (
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/quota"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/region"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal/zone"
)

// GreenNodePortalWorkflowParams contains parameters for the portal workflow.
type GreenNodePortalWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodePortalWorkflowResult contains the result of the portal workflow.
type GreenNodePortalWorkflowResult struct {
	RegionCount int
	QuotaCount  int
	ZoneCount   int
}

// GreenNodePortalWorkflow orchestrates GreenNode portal ingestion.
func GreenNodePortalWorkflow(ctx workflow.Context, params GreenNodePortalWorkflowParams) (*GreenNodePortalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodePortalWorkflow", "region", params.Region)

	result := &GreenNodePortalWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Regions
	var regionResult region.GreenNodePortalRegionWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, region.GreenNodePortalRegionWorkflow, region.GreenNodePortalRegionWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &regionResult)
	if err != nil {
		logger.Error("Failed to ingest regions", "error", err)
	} else {
		result.RegionCount = regionResult.RegionCount
	}

	// Quotas
	var quotaResult quota.GreenNodePortalQuotaWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, quota.GreenNodePortalQuotaWorkflow, quota.GreenNodePortalQuotaWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &quotaResult)
	if err != nil {
		logger.Error("Failed to ingest quotas", "error", err)
	} else {
		result.QuotaCount = quotaResult.QuotaCount
	}

	// Zones
	var zoneResult zone.GreenNodePortalZoneWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, zone.GreenNodePortalZoneWorkflow, zone.GreenNodePortalZoneWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &zoneResult)
	if err != nil {
		logger.Error("Failed to ingest zones", "error", err)
	} else {
		result.ZoneCount = zoneResult.ZoneCount
	}

	logger.Info("Completed GreenNodePortalWorkflow",
		"regionCount", result.RegionCount,
		"quotaCount", result.QuotaCount,
		"zoneCount", result.ZoneCount,
	)

	return result, nil
}
