package greennode

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest"
)

// GreenNodeInventoryWorkflowResult contains the result of GreenNode inventory collection.
type GreenNodeInventoryWorkflowResult struct {
	// Portal
	RegionCount int
	QuotaCount  int
	ZoneCount   int

	// Compute
	ServerCount      int
	SSHKeyCount      int
	ServerGroupCount int
	OSImageCount     int
	UserImageCount   int

	// Network
	SecgroupCount     int
	EndpointCount     int
	VPCCount          int
	SubnetCount       int
	RouteTableCount   int
	PeeringCount      int
	InterconnectCount int

	// Volume
	BlockVolumeCount    int
	VolumeTypeCount     int
	VolumeTypeZoneCount int

	// Load Balancer
	LBCount          int
	CertificateCount int
	LBPackageCount   int

	// GLB
	GLBCount       int
	GLBPackageCount int
	GLBRegionCount int

	// DNS
	HostedZoneCount int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*GreenNodeInventoryWorkflowResult, any)

// GreenNodeInventoryWorkflow orchestrates GreenNode inventory collection.
func GreenNodeInventoryWorkflow(ctx workflow.Context) (*GreenNodeInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeInventoryWorkflow")

	result := &GreenNodeInventoryWorkflowResult{}

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
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

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Discover projects for the first region (used for global services)
	var firstProjectID string
	{
		var projectsResult DiscoverProjectsResult
		err = workflow.ExecuteActivity(activityCtx, DiscoverProjectsActivity, DiscoverProjectsParams{
			Region: firstRegion,
		}).Get(ctx, &projectsResult)
		if err != nil {
			logger.Error("Failed to discover projects for first region", "error", err, "region", firstRegion)
			return nil, err
		}
		if len(projectsResult.ProjectIDs) == 0 {
			logger.Warn("No GreenNode projects discovered", "region", firstRegion)
			return result, nil
		}
		firstProjectID = projectsResult.ProjectIDs[0]
	}

	services := ingest.Services("greennode")

	// Global services (portal, glb, dns) — run once using first project/region
	for _, svc := range services {
		if svc.Scope != ingest.ScopeGlobal {
			continue
		}
		res := svc.NewResult()
		err = workflow.ExecuteChildWorkflow(childCtx, svc.Workflow,
			svc.NewParams(firstProjectID, firstRegion)).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	// Per-region services: discover projects per region, then run per-project services.
	for _, region := range discoverResult.Regions {
		var projectsResult DiscoverProjectsResult
		err = workflow.ExecuteActivity(activityCtx, DiscoverProjectsActivity, DiscoverProjectsParams{
			Region: region,
		}).Get(ctx, &projectsResult)
		if err != nil {
			logger.Error("Failed to discover projects for region", "error", err, "region", region)
			continue
		}

		if len(projectsResult.ProjectIDs) == 0 {
			logger.Warn("No projects discovered for region, skipping", "region", region)
			continue
		}

		logger.Info("Discovered projects for region", "region", region, "count", len(projectsResult.ProjectIDs), "projectIDs", projectsResult.ProjectIDs)

		for _, projectID := range projectsResult.ProjectIDs {
			for _, svc := range services {
				if svc.Scope != ingest.ScopeRegional {
					continue
				}
				res := svc.NewResult()
				err = workflow.ExecuteChildWorkflow(childCtx, svc.Workflow,
					svc.NewParams(projectID, region)).Get(ctx, res)
				if err != nil {
					logger.Error("Failed ingestion", "service", svc.Name, "error", err,
						"projectID", projectID, "region", region)
				} else {
					svc.Aggregate.(aggregateFunc)(result, res)
				}
			}
		}
	}

	logger.Info("Completed GreenNodeInventoryWorkflow",
		"regionCount", result.RegionCount,
		"quotaCount", result.QuotaCount,
		"zoneCount", result.ZoneCount,
		"serverCount", result.ServerCount,
		"sshKeyCount", result.SSHKeyCount,
		"serverGroupCount", result.ServerGroupCount,
		"osImageCount", result.OSImageCount,
		"userImageCount", result.UserImageCount,
		"secgroupCount", result.SecgroupCount,
		"endpointCount", result.EndpointCount,
		"vpcCount", result.VPCCount,
		"subnetCount", result.SubnetCount,
		"routeTableCount", result.RouteTableCount,
		"peeringCount", result.PeeringCount,
		"interconnectCount", result.InterconnectCount,
		"blockVolumeCount", result.BlockVolumeCount,
		"volumeTypeCount", result.VolumeTypeCount,
		"volumeTypeZoneCount", result.VolumeTypeZoneCount,
		"lbCount", result.LBCount,
		"certificateCount", result.CertificateCount,
		"lbPackageCount", result.LBPackageCount,
		"glbCount", result.GLBCount,
		"glbPackageCount", result.GLBPackageCount,
		"glbRegionCount", result.GLBRegionCount,
		"hostedZoneCount", result.HostedZoneCount,
	)

	return result, nil
}
