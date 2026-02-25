package greennode

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/dns"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/loadbalancer"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/portal"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume"
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
	GLBCount          int
	GLBPackageCount   int
	GLBRegionCount    int

	// DNS
	HostedZoneCount int
}

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
	disabled := discoverResult.DisabledServices

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Discover projects for the first region (used for portal, GLB, DNS which are global)
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

	// Portal (regions + quotas + zones) - run once using first region and first project
	if ingest.ServiceDisabled(disabled, "portal") {
		logger.Info("Skipping disabled service", "service", "portal")
	} else {
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
			result.ZoneCount = portalResult.ZoneCount
		}
	}

	// Per-region services: discover projects per region, then run per-project services.
	// Different regions have different project IDs (e.g. VNetwork endpoint API is region-scoped).
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
			// Compute
			if !ingest.ServiceDisabled(disabled, "compute") {
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
					result.OSImageCount += computeResult.OSImageCount
					result.UserImageCount += computeResult.UserImageCount
				}
			}

			// Network
			if !ingest.ServiceDisabled(disabled, "network") {
				var networkResult network.GreenNodeNetworkWorkflowResult
				err = workflow.ExecuteChildWorkflow(childCtx, network.GreenNodeNetworkWorkflow, network.GreenNodeNetworkWorkflowParams{
					ProjectID: projectID,
					Region:    region,
				}).Get(ctx, &networkResult)
				if err != nil {
					logger.Error("Failed network ingestion", "error", err, "projectID", projectID, "region", region)
				} else {
					result.SecgroupCount += networkResult.SecgroupCount
					result.EndpointCount += networkResult.EndpointCount
					result.VPCCount += networkResult.VPCCount
					result.SubnetCount += networkResult.SubnetCount
					result.RouteTableCount += networkResult.RouteTableCount
					result.PeeringCount += networkResult.PeeringCount
					result.InterconnectCount += networkResult.InterconnectCount
				}
			}

			// Volume
			if !ingest.ServiceDisabled(disabled, "volume") {
				var volumeResult volume.GreenNodeVolumeWorkflowResult
				err = workflow.ExecuteChildWorkflow(childCtx, volume.GreenNodeVolumeWorkflow, volume.GreenNodeVolumeWorkflowParams{
					ProjectID: projectID,
					Region:    region,
				}).Get(ctx, &volumeResult)
				if err != nil {
					logger.Error("Failed volume ingestion", "error", err, "projectID", projectID, "region", region)
				} else {
					result.BlockVolumeCount += volumeResult.BlockVolumeCount
					result.VolumeTypeCount += volumeResult.VolumeTypeCount
					result.VolumeTypeZoneCount += volumeResult.VolumeTypeZoneCount
				}
			}

			// Load Balancer
			if !ingest.ServiceDisabled(disabled, "loadbalancer") {
				var lbResult loadbalancer.GreenNodeLoadBalancerWorkflowResult
				err = workflow.ExecuteChildWorkflow(childCtx, loadbalancer.GreenNodeLoadBalancerWorkflow, loadbalancer.GreenNodeLoadBalancerWorkflowParams{
					ProjectID: projectID,
					Region:    region,
				}).Get(ctx, &lbResult)
				if err != nil {
					logger.Error("Failed load balancer ingestion", "error", err, "projectID", projectID, "region", region)
				} else {
					result.LBCount += lbResult.LBCount
					result.CertificateCount += lbResult.CertificateCount
					result.LBPackageCount += lbResult.PackageCount
				}
			}
		}
	}

	// GLB - run once globally
	if ingest.ServiceDisabled(disabled, "glb") {
		logger.Info("Skipping disabled service", "service", "glb")
	} else {
		var glbResult glb.GreenNodeGLBWorkflowResult
		err = workflow.ExecuteChildWorkflow(childCtx, glb.GreenNodeGLBWorkflow, glb.GreenNodeGLBWorkflowParams{
			ProjectID: firstProjectID,
			Region:    firstRegion,
		}).Get(ctx, &glbResult)
		if err != nil {
			logger.Error("Failed GLB ingestion", "error", err)
		} else {
			result.GLBCount = glbResult.GLBCount
			result.GLBPackageCount = glbResult.PackageCount
			result.GLBRegionCount = glbResult.RegionCount
		}
	}

	// DNS - run once globally
	if ingest.ServiceDisabled(disabled, "dns") {
		logger.Info("Skipping disabled service", "service", "dns")
	} else {
		var dnsResult dns.GreenNodeDNSWorkflowResult
		err = workflow.ExecuteChildWorkflow(childCtx, dns.GreenNodeDNSWorkflow, dns.GreenNodeDNSWorkflowParams{
			ProjectID: firstProjectID,
			Region:    firstRegion,
		}).Get(ctx, &dnsResult)
		if err != nil {
			logger.Error("Failed DNS ingestion", "error", err)
		} else {
			result.HostedZoneCount = dnsResult.HostedZoneCount
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
