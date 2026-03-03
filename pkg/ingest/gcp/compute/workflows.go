package compute

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/address"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/backendservice"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/disk"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/firewall"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globaladdress"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/healthcheck"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/image"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instance"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instancegroup"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/interconnect"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/neg"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/negendpoint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/network"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/packetmirroring"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/projectmetadata"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/router"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/securitypolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/snapshot"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/sslpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/subnetwork"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targethttpproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targethttpsproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetinstance"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetpool"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetsslproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targettcpproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetvpngateway"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/urlmap"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/vpngateway"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/vpntunnel"
)

// GCPComputeWorkflowParams contains parameters for the compute workflow.
type GCPComputeWorkflowParams struct {
	ProjectID string
}

// GCPComputeWorkflowResult contains the result of the compute workflow.
type GCPComputeWorkflowResult struct {
	ProjectID                 string
	InstanceCount             int
	DiskCount                 int
	NetworkCount              int
	SubnetworkCount           int
	InstanceGroupCount        int
	TargetInstanceCount       int
	AddressCount              int
	GlobalAddressCount        int
	SnapshotCount             int
	ImageCount                int
	HealthCheckCount          int
	ForwardingRuleCount       int
	GlobalForwardingRuleCount int
	VpnGatewayCount           int
	TargetVpnGatewayCount     int
	VpnTunnelCount            int
	TargetHttpProxyCount      int
	TargetTcpProxyCount       int
	TargetSslProxyCount       int
	TargetHttpsProxyCount     int
	UrlMapCount               int
	TargetPoolCount           int
	NegCount                  int
	NegEndpointCount          int
	BackendServiceCount       int
	FirewallCount             int
	SslPolicyCount            int
	RouterCount               int
	SecurityPolicyCount       int
	InterconnectCount         int
	PacketMirroringCount      int
	ProjectMetadataCount      int
}

// GCPComputeWorkflow ingests all GCP Compute Engine resources for a single project.
// All independent resources run in parallel. NEGEndpoint runs after NEG (it queries NEGs from DB).
func GCPComputeWorkflow(ctx workflow.Context, params GCPComputeWorkflowParams) (*GCPComputeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPComputeWorkflow", "projectID", params.ProjectID)

	// Child workflow options
	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 30 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	result := &GCPComputeWorkflowResult{
		ProjectID: params.ProjectID,
	}

	// Phase 1: Launch all independent resources in parallel
	instanceFuture := workflow.ExecuteChildWorkflow(childCtx, instance.GCPComputeInstanceWorkflow,
		instance.GCPComputeInstanceWorkflowParams{ProjectID: params.ProjectID})

	diskFuture := workflow.ExecuteChildWorkflow(childCtx, disk.GCPComputeDiskWorkflow,
		disk.GCPComputeDiskWorkflowParams{ProjectID: params.ProjectID})

	networkFuture := workflow.ExecuteChildWorkflow(childCtx, network.GCPComputeNetworkWorkflow,
		network.GCPComputeNetworkWorkflowParams{ProjectID: params.ProjectID})

	subnetworkFuture := workflow.ExecuteChildWorkflow(childCtx, subnetwork.GCPComputeSubnetworkWorkflow,
		subnetwork.GCPComputeSubnetworkWorkflowParams{ProjectID: params.ProjectID})

	instanceGroupFuture := workflow.ExecuteChildWorkflow(childCtx, instancegroup.GCPComputeInstanceGroupWorkflow,
		instancegroup.GCPComputeInstanceGroupWorkflowParams{ProjectID: params.ProjectID})

	targetInstanceFuture := workflow.ExecuteChildWorkflow(childCtx, targetinstance.GCPComputeTargetInstanceWorkflow,
		targetinstance.GCPComputeTargetInstanceWorkflowParams{ProjectID: params.ProjectID})

	addressFuture := workflow.ExecuteChildWorkflow(childCtx, address.GCPComputeAddressWorkflow,
		address.GCPComputeAddressWorkflowParams{ProjectID: params.ProjectID})

	globalAddressFuture := workflow.ExecuteChildWorkflow(childCtx, globaladdress.GCPComputeGlobalAddressWorkflow,
		globaladdress.GCPComputeGlobalAddressWorkflowParams{ProjectID: params.ProjectID})

	forwardingRuleFuture := workflow.ExecuteChildWorkflow(childCtx, forwardingrule.GCPComputeForwardingRuleWorkflow,
		forwardingrule.GCPComputeForwardingRuleWorkflowParams{ProjectID: params.ProjectID})

	globalForwardingRuleFuture := workflow.ExecuteChildWorkflow(childCtx, globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflow,
		globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflowParams{ProjectID: params.ProjectID})

	snapshotFuture := workflow.ExecuteChildWorkflow(childCtx, snapshot.GCPComputeSnapshotWorkflow,
		snapshot.GCPComputeSnapshotWorkflowParams{ProjectID: params.ProjectID})

	imageFuture := workflow.ExecuteChildWorkflow(childCtx, image.GCPComputeImageWorkflow,
		image.GCPComputeImageWorkflowParams{ProjectID: params.ProjectID})

	healthCheckFuture := workflow.ExecuteChildWorkflow(childCtx, healthcheck.GCPComputeHealthCheckWorkflow,
		healthcheck.GCPComputeHealthCheckWorkflowParams{ProjectID: params.ProjectID})

	vpnGatewayFuture := workflow.ExecuteChildWorkflow(childCtx, vpngateway.GCPComputeVpnGatewayWorkflow,
		vpngateway.GCPComputeVpnGatewayWorkflowParams{ProjectID: params.ProjectID})

	targetVpnGatewayFuture := workflow.ExecuteChildWorkflow(childCtx, targetvpngateway.GCPComputeTargetVpnGatewayWorkflow,
		targetvpngateway.GCPComputeTargetVpnGatewayWorkflowParams{ProjectID: params.ProjectID})

	vpnTunnelFuture := workflow.ExecuteChildWorkflow(childCtx, vpntunnel.GCPComputeVpnTunnelWorkflow,
		vpntunnel.GCPComputeVpnTunnelWorkflowParams{ProjectID: params.ProjectID})

	targetHttpProxyFuture := workflow.ExecuteChildWorkflow(childCtx, targethttpproxy.GCPComputeTargetHttpProxyWorkflow,
		targethttpproxy.GCPComputeTargetHttpProxyWorkflowParams{ProjectID: params.ProjectID})

	targetTcpProxyFuture := workflow.ExecuteChildWorkflow(childCtx, targettcpproxy.GCPComputeTargetTcpProxyWorkflow,
		targettcpproxy.GCPComputeTargetTcpProxyWorkflowParams{ProjectID: params.ProjectID})

	targetSslProxyFuture := workflow.ExecuteChildWorkflow(childCtx, targetsslproxy.GCPComputeTargetSslProxyWorkflow,
		targetsslproxy.GCPComputeTargetSslProxyWorkflowParams{ProjectID: params.ProjectID})

	targetHttpsProxyFuture := workflow.ExecuteChildWorkflow(childCtx, targethttpsproxy.GCPComputeTargetHttpsProxyWorkflow,
		targethttpsproxy.GCPComputeTargetHttpsProxyWorkflowParams{ProjectID: params.ProjectID})

	urlMapFuture := workflow.ExecuteChildWorkflow(childCtx, urlmap.GCPComputeUrlMapWorkflow,
		urlmap.GCPComputeUrlMapWorkflowParams{ProjectID: params.ProjectID})

	targetPoolFuture := workflow.ExecuteChildWorkflow(childCtx, targetpool.GCPComputeTargetPoolWorkflow,
		targetpool.GCPComputeTargetPoolWorkflowParams{ProjectID: params.ProjectID})

	negFuture := workflow.ExecuteChildWorkflow(childCtx, neg.GCPComputeNegWorkflow,
		neg.GCPComputeNegWorkflowParams{ProjectID: params.ProjectID})

	backendServiceFuture := workflow.ExecuteChildWorkflow(childCtx, backendservice.GCPComputeBackendServiceWorkflow,
		backendservice.GCPComputeBackendServiceWorkflowParams{ProjectID: params.ProjectID})

	firewallFuture := workflow.ExecuteChildWorkflow(childCtx, firewall.GCPComputeFirewallWorkflow,
		firewall.GCPComputeFirewallWorkflowParams{ProjectID: params.ProjectID})

	sslPolicyFuture := workflow.ExecuteChildWorkflow(childCtx, sslpolicy.GCPComputeSslPolicyWorkflow,
		sslpolicy.GCPComputeSslPolicyWorkflowParams{ProjectID: params.ProjectID})

	routerFuture := workflow.ExecuteChildWorkflow(childCtx, router.GCPComputeRouterWorkflow,
		router.GCPComputeRouterWorkflowParams{ProjectID: params.ProjectID})

	securityPolicyFuture := workflow.ExecuteChildWorkflow(childCtx, securitypolicy.GCPComputeSecurityPolicyWorkflow,
		securitypolicy.GCPComputeSecurityPolicyWorkflowParams{ProjectID: params.ProjectID})

	interconnectFuture := workflow.ExecuteChildWorkflow(childCtx, interconnect.GCPComputeInterconnectWorkflow,
		interconnect.GCPComputeInterconnectWorkflowParams{ProjectID: params.ProjectID})

	packetMirroringFuture := workflow.ExecuteChildWorkflow(childCtx, packetmirroring.GCPComputePacketMirroringWorkflow,
		packetmirroring.GCPComputePacketMirroringWorkflowParams{ProjectID: params.ProjectID})

	projectMetadataFuture := workflow.ExecuteChildWorkflow(childCtx, projectmetadata.GCPComputeProjectMetadataWorkflow,
		projectmetadata.GCPComputeProjectMetadataWorkflowParams{ProjectID: params.ProjectID})

	// Phase 2: Collect all phase 1 results (continue on error to avoid killing parallel workflows)
	var errs []error

	var instanceResult instance.GCPComputeInstanceWorkflowResult
	if err := instanceFuture.Get(ctx, &instanceResult); err != nil {
		logger.Error("Failed to ingest instances", "error", err)
		errs = append(errs, err)
	} else {
		result.InstanceCount = instanceResult.InstanceCount
	}

	var diskResult disk.GCPComputeDiskWorkflowResult
	if err := diskFuture.Get(ctx, &diskResult); err != nil {
		logger.Error("Failed to ingest disks", "error", err)
		errs = append(errs, err)
	} else {
		result.DiskCount = diskResult.DiskCount
	}

	var networkResult network.GCPComputeNetworkWorkflowResult
	if err := networkFuture.Get(ctx, &networkResult); err != nil {
		logger.Error("Failed to ingest networks", "error", err)
		errs = append(errs, err)
	} else {
		result.NetworkCount = networkResult.NetworkCount
	}

	var subnetworkResult subnetwork.GCPComputeSubnetworkWorkflowResult
	if err := subnetworkFuture.Get(ctx, &subnetworkResult); err != nil {
		logger.Error("Failed to ingest subnetworks", "error", err)
		errs = append(errs, err)
	} else {
		result.SubnetworkCount = subnetworkResult.SubnetworkCount
	}

	var instanceGroupResult instancegroup.GCPComputeInstanceGroupWorkflowResult
	if err := instanceGroupFuture.Get(ctx, &instanceGroupResult); err != nil {
		logger.Error("Failed to ingest instance groups", "error", err)
		errs = append(errs, err)
	} else {
		result.InstanceGroupCount = instanceGroupResult.InstanceGroupCount
	}

	var targetInstanceResult targetinstance.GCPComputeTargetInstanceWorkflowResult
	if err := targetInstanceFuture.Get(ctx, &targetInstanceResult); err != nil {
		logger.Error("Failed to ingest target instances", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetInstanceCount = targetInstanceResult.TargetInstanceCount
	}

	var addressResult address.GCPComputeAddressWorkflowResult
	if err := addressFuture.Get(ctx, &addressResult); err != nil {
		logger.Error("Failed to ingest addresses", "error", err)
		errs = append(errs, err)
	} else {
		result.AddressCount = addressResult.AddressCount
	}

	var globalAddressResult globaladdress.GCPComputeGlobalAddressWorkflowResult
	if err := globalAddressFuture.Get(ctx, &globalAddressResult); err != nil {
		logger.Error("Failed to ingest global addresses", "error", err)
		errs = append(errs, err)
	} else {
		result.GlobalAddressCount = globalAddressResult.GlobalAddressCount
	}

	var forwardingRuleResult forwardingrule.GCPComputeForwardingRuleWorkflowResult
	if err := forwardingRuleFuture.Get(ctx, &forwardingRuleResult); err != nil {
		logger.Error("Failed to ingest forwarding rules", "error", err)
		errs = append(errs, err)
	} else {
		result.ForwardingRuleCount = forwardingRuleResult.ForwardingRuleCount
	}

	var globalForwardingRuleResult globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflowResult
	if err := globalForwardingRuleFuture.Get(ctx, &globalForwardingRuleResult); err != nil {
		logger.Error("Failed to ingest global forwarding rules", "error", err)
		errs = append(errs, err)
	} else {
		result.GlobalForwardingRuleCount = globalForwardingRuleResult.GlobalForwardingRuleCount
	}

	var snapshotResult snapshot.GCPComputeSnapshotWorkflowResult
	if err := snapshotFuture.Get(ctx, &snapshotResult); err != nil {
		logger.Error("Failed to ingest snapshots", "error", err)
		errs = append(errs, err)
	} else {
		result.SnapshotCount = snapshotResult.SnapshotCount
	}

	var imageResult image.GCPComputeImageWorkflowResult
	if err := imageFuture.Get(ctx, &imageResult); err != nil {
		logger.Error("Failed to ingest images", "error", err)
		errs = append(errs, err)
	} else {
		result.ImageCount = imageResult.ImageCount
	}

	var healthCheckResult healthcheck.GCPComputeHealthCheckWorkflowResult
	if err := healthCheckFuture.Get(ctx, &healthCheckResult); err != nil {
		logger.Error("Failed to ingest health checks", "error", err)
		errs = append(errs, err)
	} else {
		result.HealthCheckCount = healthCheckResult.HealthCheckCount
	}

	var vpnGatewayResult vpngateway.GCPComputeVpnGatewayWorkflowResult
	if err := vpnGatewayFuture.Get(ctx, &vpnGatewayResult); err != nil {
		logger.Error("Failed to ingest vpn gateways", "error", err)
		errs = append(errs, err)
	} else {
		result.VpnGatewayCount = vpnGatewayResult.VpnGatewayCount
	}

	var targetVpnGatewayResult targetvpngateway.GCPComputeTargetVpnGatewayWorkflowResult
	if err := targetVpnGatewayFuture.Get(ctx, &targetVpnGatewayResult); err != nil {
		logger.Error("Failed to ingest target vpn gateways", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetVpnGatewayCount = targetVpnGatewayResult.TargetVpnGatewayCount
	}

	var vpnTunnelResult vpntunnel.GCPComputeVpnTunnelWorkflowResult
	if err := vpnTunnelFuture.Get(ctx, &vpnTunnelResult); err != nil {
		logger.Error("Failed to ingest vpn tunnels", "error", err)
		errs = append(errs, err)
	} else {
		result.VpnTunnelCount = vpnTunnelResult.VpnTunnelCount
	}

	var targetHttpProxyResult targethttpproxy.GCPComputeTargetHttpProxyWorkflowResult
	if err := targetHttpProxyFuture.Get(ctx, &targetHttpProxyResult); err != nil {
		logger.Error("Failed to ingest target HTTP proxies", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetHttpProxyCount = targetHttpProxyResult.TargetHttpProxyCount
	}

	var targetTcpProxyResult targettcpproxy.GCPComputeTargetTcpProxyWorkflowResult
	if err := targetTcpProxyFuture.Get(ctx, &targetTcpProxyResult); err != nil {
		logger.Error("Failed to ingest target TCP proxies", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetTcpProxyCount = targetTcpProxyResult.TargetTcpProxyCount
	}

	var targetSslProxyResult targetsslproxy.GCPComputeTargetSslProxyWorkflowResult
	if err := targetSslProxyFuture.Get(ctx, &targetSslProxyResult); err != nil {
		logger.Error("Failed to ingest target SSL proxies", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetSslProxyCount = targetSslProxyResult.TargetSslProxyCount
	}

	var targetHttpsProxyResult targethttpsproxy.GCPComputeTargetHttpsProxyWorkflowResult
	if err := targetHttpsProxyFuture.Get(ctx, &targetHttpsProxyResult); err != nil {
		logger.Error("Failed to ingest target HTTPS proxies", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetHttpsProxyCount = targetHttpsProxyResult.TargetHttpsProxyCount
	}

	var urlMapResult urlmap.GCPComputeUrlMapWorkflowResult
	if err := urlMapFuture.Get(ctx, &urlMapResult); err != nil {
		logger.Error("Failed to ingest URL maps", "error", err)
		errs = append(errs, err)
	} else {
		result.UrlMapCount = urlMapResult.UrlMapCount
	}

	var targetPoolResult targetpool.GCPComputeTargetPoolWorkflowResult
	if err := targetPoolFuture.Get(ctx, &targetPoolResult); err != nil {
		logger.Error("Failed to ingest target pools", "error", err)
		errs = append(errs, err)
	} else {
		result.TargetPoolCount = targetPoolResult.TargetPoolCount
	}

	var negResult neg.GCPComputeNegWorkflowResult
	if err := negFuture.Get(ctx, &negResult); err != nil {
		logger.Error("Failed to ingest NEGs", "error", err)
		errs = append(errs, err)
	} else {
		result.NegCount = negResult.NegCount
	}

	var backendServiceResult backendservice.GCPComputeBackendServiceWorkflowResult
	if err := backendServiceFuture.Get(ctx, &backendServiceResult); err != nil {
		logger.Error("Failed to ingest backend services", "error", err)
		errs = append(errs, err)
	} else {
		result.BackendServiceCount = backendServiceResult.BackendServiceCount
	}

	var firewallResult firewall.GCPComputeFirewallWorkflowResult
	if err := firewallFuture.Get(ctx, &firewallResult); err != nil {
		logger.Error("Failed to ingest firewalls", "error", err)
		errs = append(errs, err)
	} else {
		result.FirewallCount = firewallResult.FirewallCount
	}

	var sslPolicyResult sslpolicy.GCPComputeSslPolicyWorkflowResult
	if err := sslPolicyFuture.Get(ctx, &sslPolicyResult); err != nil {
		logger.Error("Failed to ingest SSL policies", "error", err)
		errs = append(errs, err)
	} else {
		result.SslPolicyCount = sslPolicyResult.SslPolicyCount
	}

	var routerResult router.GCPComputeRouterWorkflowResult
	if err := routerFuture.Get(ctx, &routerResult); err != nil {
		logger.Error("Failed to ingest routers", "error", err)
		errs = append(errs, err)
	} else {
		result.RouterCount = routerResult.RouterCount
	}

	var securityPolicyResult securitypolicy.GCPComputeSecurityPolicyWorkflowResult
	if err := securityPolicyFuture.Get(ctx, &securityPolicyResult); err != nil {
		logger.Error("Failed to ingest security policies", "error", err)
		errs = append(errs, err)
	} else {
		result.SecurityPolicyCount = securityPolicyResult.SecurityPolicyCount
	}

	var interconnectResult interconnect.GCPComputeInterconnectWorkflowResult
	if err := interconnectFuture.Get(ctx, &interconnectResult); err != nil {
		logger.Error("Failed to ingest interconnects", "error", err)
		errs = append(errs, err)
	} else {
		result.InterconnectCount = interconnectResult.InterconnectCount
	}

	var packetMirroringResult packetmirroring.GCPComputePacketMirroringWorkflowResult
	if err := packetMirroringFuture.Get(ctx, &packetMirroringResult); err != nil {
		logger.Error("Failed to ingest packet mirrorings", "error", err)
		errs = append(errs, err)
	} else {
		result.PacketMirroringCount = packetMirroringResult.PacketMirroringCount
	}

	var projectMetadataResult projectmetadata.GCPComputeProjectMetadataWorkflowResult
	if err := projectMetadataFuture.Get(ctx, &projectMetadataResult); err != nil {
		logger.Error("Failed to ingest project metadata", "error", err)
		errs = append(errs, err)
	} else {
		result.ProjectMetadataCount = projectMetadataResult.MetadataCount
	}

	// Phase 3: NEG endpoints depend on NEGs being in DB
	var negEndpointResult negendpoint.GCPComputeNegEndpointWorkflowResult
	if err := workflow.ExecuteChildWorkflow(childCtx, negendpoint.GCPComputeNegEndpointWorkflow,
		negendpoint.GCPComputeNegEndpointWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &negEndpointResult); err != nil {
		logger.Error("Failed to ingest NEG endpoints", "error", err)
		errs = append(errs, err)
	} else {
		result.NegEndpointCount = negEndpointResult.NegEndpointCount
	}

	if len(errs) > 0 {
		logger.Warn("GCPComputeWorkflow completed with errors",
			"projectID", params.ProjectID, "errorCount", len(errs))
	}

	logger.Info("Completed GCPComputeWorkflow",
		"projectID", params.ProjectID,
		"instanceCount", result.InstanceCount,
		"diskCount", result.DiskCount,
		"networkCount", result.NetworkCount,
		"subnetworkCount", result.SubnetworkCount,
		"instanceGroupCount", result.InstanceGroupCount,
		"targetInstanceCount", result.TargetInstanceCount,
		"addressCount", result.AddressCount,
		"globalAddressCount", result.GlobalAddressCount,
		"forwardingRuleCount", result.ForwardingRuleCount,
		"globalForwardingRuleCount", result.GlobalForwardingRuleCount,
		"snapshotCount", result.SnapshotCount,
		"imageCount", result.ImageCount,
		"healthCheckCount", result.HealthCheckCount,
		"vpnGatewayCount", result.VpnGatewayCount,
		"targetVpnGatewayCount", result.TargetVpnGatewayCount,
		"vpnTunnelCount", result.VpnTunnelCount,
		"targetHttpProxyCount", result.TargetHttpProxyCount,
		"targetTcpProxyCount", result.TargetTcpProxyCount,
		"targetSslProxyCount", result.TargetSslProxyCount,
		"targetHttpsProxyCount", result.TargetHttpsProxyCount,
		"urlMapCount", result.UrlMapCount,
		"targetPoolCount", result.TargetPoolCount,
		"negCount", result.NegCount,
		"negEndpointCount", result.NegEndpointCount,
		"backendServiceCount", result.BackendServiceCount,
		"firewallCount", result.FirewallCount,
		"sslPolicyCount", result.SslPolicyCount,
		"routerCount", result.RouterCount,
		"securityPolicyCount", result.SecurityPolicyCount,
		"interconnectCount", result.InterconnectCount,
		"packetMirroringCount", result.PacketMirroringCount,
		"projectMetadataCount", result.ProjectMetadataCount,
	)

	return result, nil
}
