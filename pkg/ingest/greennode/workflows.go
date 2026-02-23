package greennode

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal"
)

// GreenNodeInventoryWorkflowResult contains the result of GreenNode inventory collection.
type GreenNodeInventoryWorkflowResult struct {
	// Portal
	RegionCount int
	QuotaCount  int

	// Compute
	ServerCount      int
	SSHKeyCount      int
	ServerGroupCount int
}

// GreenNodeInventoryWorkflow orchestrates GreenNode inventory collection.
func GreenNodeInventoryWorkflow(ctx workflow.Context) (*GreenNodeInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeInventoryWorkflow")

	result := &GreenNodeInventoryWorkflowResult{}

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	// Discover configured regions
	var discoverResult DiscoverRegionsResult
	err := workflow.ExecuteActivity(activityCtx, DiscoverRegionsActivity, DiscoverRegionsParams{}).Get(ctx, &discoverResult)
	if err != nil {
		logger.Error("Failed to discover regions", "error", err)
		return nil, err
	}

	if len(discoverResult.Regions) == 0 {
		logger.Warn("No GreenNode regions configured")
		return result, nil
	}

	firstRegion := discoverResult.Regions[0]

	// Discover projects (from config or via Portal V1 API)
	var projectsResult DiscoverProjectsResult
	err = workflow.ExecuteActivity(activityCtx, DiscoverProjectsActivity, DiscoverProjectsParams{
		Region: firstRegion,
	}).Get(ctx, &projectsResult)
	if err != nil {
		logger.Error("Failed to discover projects", "error", err)
		return nil, err
	}

	if len(projectsResult.ProjectIDs) == 0 {
		logger.Warn("No GreenNode projects discovered")
		return result, nil
	}

	logger.Info("Discovered projects", "count", len(projectsResult.ProjectIDs), "projectIDs", projectsResult.ProjectIDs)

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Use first project for portal (regions + quotas are project-scoped)
	firstProjectID := projectsResult.ProjectIDs[0]

	// Portal (regions + quotas) - run once using first region and first project
	var portalResult portal.GreenNodePortalWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, portal.GreenNodePortalWorkflow, portal.GreenNodePortalWorkflowParams{
		ProjectID: firstProjectID,
		Region:    firstRegion,
	}).Get(ctx, &portalResult)
	if err != nil {
		logger.Error("Failed portal ingestion", "error", err)
	} else {
		result.RegionCount = portalResult.RegionCount
		result.QuotaCount = portalResult.QuotaCount
	}

	// Compute - run per (project, region)
	for _, projectID := range projectsResult.ProjectIDs {
		for _, region := range discoverResult.Regions {
			var computeResult compute.GreenNodeComputeWorkflowResult
			err = workflow.ExecuteChildWorkflow(childCtx, compute.GreenNodeComputeWorkflow, compute.GreenNodeComputeWorkflowParams{
				ProjectID: projectID,
				Region:    region,
			}).Get(ctx, &computeResult)
			if err != nil {
				logger.Error("Failed compute ingestion", "error", err, "projectID", projectID, "region", region)
			} else {
				result.ServerCount += computeResult.ServerCount
				result.SSHKeyCount += computeResult.SSHKeyCount
				result.ServerGroupCount += computeResult.ServerGroupCount
			}
		}
	}

	logger.Info("Completed GreenNodeInventoryWorkflow",
		"regionCount", result.RegionCount,
		"quotaCount", result.QuotaCount,
		"serverCount", result.ServerCount,
		"sshKeyCount", result.SSHKeyCount,
		"serverGroupCount", result.ServerGroupCount,
	)

	return result, nil
}
