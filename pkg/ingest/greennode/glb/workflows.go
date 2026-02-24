package glb

import (
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbpackage"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbregion"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbresource"
)

// GreenNodeGLBWorkflowParams contains parameters for the GLB workflow.
type GreenNodeGLBWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeGLBWorkflowResult contains the result of the GLB workflow.
type GreenNodeGLBWorkflowResult struct {
	GLBCount     int
	PackageCount int
	RegionCount  int
}

// GreenNodeGLBWorkflow orchestrates GreenNode GLB ingestion.
func GreenNodeGLBWorkflow(ctx workflow.Context, params GreenNodeGLBWorkflowParams) (*GreenNodeGLBWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeGLBWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeGLBWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Global Load Balancers
	var glbResult glbresource.GreenNodeGLBGlobalLoadBalancerWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, glbresource.GreenNodeGLBGlobalLoadBalancerWorkflow, glbresource.GreenNodeGLBGlobalLoadBalancerWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &glbResult)
	if err != nil {
		logger.Error("Failed to ingest global load balancers", "error", err)
	} else {
		result.GLBCount = glbResult.GLBCount
	}

	// Global Packages
	var pkgResult glbpackage.GreenNodeGLBGlobalPackageWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, glbpackage.GreenNodeGLBGlobalPackageWorkflow, glbpackage.GreenNodeGLBGlobalPackageWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &pkgResult)
	if err != nil {
		logger.Error("Failed to ingest global packages", "error", err)
	} else {
		result.PackageCount = pkgResult.PackageCount
	}

	// Global Regions
	var regionResult glbregion.GreenNodeGLBGlobalRegionWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, glbregion.GreenNodeGLBGlobalRegionWorkflow, glbregion.GreenNodeGLBGlobalRegionWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &regionResult)
	if err != nil {
		logger.Error("Failed to ingest global regions", "error", err)
	} else {
		result.RegionCount = regionResult.RegionCount
	}

	logger.Info("Completed GreenNodeGLBWorkflow",
		"glbCount", result.GLBCount,
		"packageCount", result.PackageCount,
		"regionCount", result.RegionCount,
	)

	return result, nil
}
