package compute

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/address"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/backendservice"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/disk"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globaladdress"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/healthcheck"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/image"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instance"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instancegroup"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/neg"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/negendpoint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/network"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/snapshot"
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
	ProjectID           string
	InstanceCount       int
	DiskCount           int
	NetworkCount        int
	SubnetworkCount     int
	InstanceGroupCount  int
	TargetInstanceCount int
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
}

// GCPComputeWorkflow ingests all GCP Compute Engine resources for a single project.
// Orchestrates child workflows - each manages its own session and client lifecycle.
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

	// Execute instance workflow
	var instanceResult instance.GCPComputeInstanceWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, instance.GCPComputeInstanceWorkflow,
		instance.GCPComputeInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceResult)
	if err != nil {
		logger.Error("Failed to ingest instances", "error", err)
		return nil, err
	}
	result.InstanceCount = instanceResult.InstanceCount

	// Execute disk workflow
	var diskResult disk.GCPComputeDiskWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, disk.GCPComputeDiskWorkflow,
		disk.GCPComputeDiskWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &diskResult)
	if err != nil {
		logger.Error("Failed to ingest disks", "error", err)
		return nil, err
	}
	result.DiskCount = diskResult.DiskCount

	// Execute network workflow
	var networkResult network.GCPComputeNetworkWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, network.GCPComputeNetworkWorkflow,
		network.GCPComputeNetworkWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &networkResult)
	if err != nil {
		logger.Error("Failed to ingest networks", "error", err)
		return nil, err
	}
	result.NetworkCount = networkResult.NetworkCount

	// Execute subnetwork workflow
	var subnetworkResult subnetwork.GCPComputeSubnetworkWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, subnetwork.GCPComputeSubnetworkWorkflow,
		subnetwork.GCPComputeSubnetworkWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &subnetworkResult)
	if err != nil {
		logger.Error("Failed to ingest subnetworks", "error", err)
		return nil, err
	}
	result.SubnetworkCount = subnetworkResult.SubnetworkCount

	// Execute instance group workflow
	var instanceGroupResult instancegroup.GCPComputeInstanceGroupWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, instancegroup.GCPComputeInstanceGroupWorkflow,
		instancegroup.GCPComputeInstanceGroupWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &instanceGroupResult)
	if err != nil {
		logger.Error("Failed to ingest instance groups", "error", err)
		return nil, err
	}
	result.InstanceGroupCount = instanceGroupResult.InstanceGroupCount

	// Execute target instance workflow
	var targetInstanceResult targetinstance.GCPComputeTargetInstanceWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targetinstance.GCPComputeTargetInstanceWorkflow,
		targetinstance.GCPComputeTargetInstanceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetInstanceResult)
	if err != nil {
		logger.Error("Failed to ingest target instances", "error", err)
		return nil, err
	}
	result.TargetInstanceCount = targetInstanceResult.TargetInstanceCount

	// Execute address workflow
	var addressResult address.GCPComputeAddressWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, address.GCPComputeAddressWorkflow,
		address.GCPComputeAddressWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &addressResult)
	if err != nil {
		logger.Error("Failed to ingest addresses", "error", err)
		return nil, err
	}
	result.AddressCount = addressResult.AddressCount

	// Execute global address workflow
	var globalAddressResult globaladdress.GCPComputeGlobalAddressWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, globaladdress.GCPComputeGlobalAddressWorkflow,
		globaladdress.GCPComputeGlobalAddressWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &globalAddressResult)
	if err != nil {
		logger.Error("Failed to ingest global addresses", "error", err)
		return nil, err
	}
	result.GlobalAddressCount = globalAddressResult.GlobalAddressCount

	// Execute forwarding rule workflow
	var forwardingRuleResult forwardingrule.GCPComputeForwardingRuleWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, forwardingrule.GCPComputeForwardingRuleWorkflow,
		forwardingrule.GCPComputeForwardingRuleWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &forwardingRuleResult)
	if err != nil {
		logger.Error("Failed to ingest forwarding rules", "error", err)
		return nil, err
	}
	result.ForwardingRuleCount = forwardingRuleResult.ForwardingRuleCount

	// Execute global forwarding rule workflow
	var globalForwardingRuleResult globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflow,
		globalforwardingrule.GCPComputeGlobalForwardingRuleWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &globalForwardingRuleResult)
	if err != nil {
		logger.Error("Failed to ingest global forwarding rules", "error", err)
		return nil, err
	}
	result.GlobalForwardingRuleCount = globalForwardingRuleResult.GlobalForwardingRuleCount

	// Execute snapshot workflow
	var snapshotResult snapshot.GCPComputeSnapshotWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, snapshot.GCPComputeSnapshotWorkflow,
		snapshot.GCPComputeSnapshotWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &snapshotResult)
	if err != nil {
		logger.Error("Failed to ingest snapshots", "error", err)
		return nil, err
	}
	result.SnapshotCount = snapshotResult.SnapshotCount

	// Execute image workflow
	var imageResult image.GCPComputeImageWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, image.GCPComputeImageWorkflow,
		image.GCPComputeImageWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &imageResult)
	if err != nil {
		logger.Error("Failed to ingest images", "error", err)
		return nil, err
	}
	result.ImageCount = imageResult.ImageCount

	// Execute health check workflow
	var healthCheckResult healthcheck.GCPComputeHealthCheckWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, healthcheck.GCPComputeHealthCheckWorkflow,
		healthcheck.GCPComputeHealthCheckWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &healthCheckResult)
	if err != nil {
		logger.Error("Failed to ingest health checks", "error", err)
		return nil, err
	}
	result.HealthCheckCount = healthCheckResult.HealthCheckCount

	// Execute VPN gateway workflow
	var vpnGatewayResult vpngateway.GCPComputeVpnGatewayWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, vpngateway.GCPComputeVpnGatewayWorkflow,
		vpngateway.GCPComputeVpnGatewayWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &vpnGatewayResult)
	if err != nil {
		logger.Error("Failed to ingest vpn gateways", "error", err)
		return nil, err
	}
	result.VpnGatewayCount = vpnGatewayResult.VpnGatewayCount

	// Execute target VPN gateway workflow
	var targetVpnGatewayResult targetvpngateway.GCPComputeTargetVpnGatewayWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targetvpngateway.GCPComputeTargetVpnGatewayWorkflow,
		targetvpngateway.GCPComputeTargetVpnGatewayWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetVpnGatewayResult)
	if err != nil {
		logger.Error("Failed to ingest target vpn gateways", "error", err)
		return nil, err
	}
	result.TargetVpnGatewayCount = targetVpnGatewayResult.TargetVpnGatewayCount

	// Execute VPN tunnel workflow
	var vpnTunnelResult vpntunnel.GCPComputeVpnTunnelWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, vpntunnel.GCPComputeVpnTunnelWorkflow,
		vpntunnel.GCPComputeVpnTunnelWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &vpnTunnelResult)
	if err != nil {
		logger.Error("Failed to ingest vpn tunnels", "error", err)
		return nil, err
	}
	result.VpnTunnelCount = vpnTunnelResult.VpnTunnelCount

	// Execute target HTTP proxy workflow
	var targetHttpProxyResult targethttpproxy.GCPComputeTargetHttpProxyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targethttpproxy.GCPComputeTargetHttpProxyWorkflow,
		targethttpproxy.GCPComputeTargetHttpProxyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetHttpProxyResult)
	if err != nil {
		logger.Error("Failed to ingest target HTTP proxies", "error", err)
		return nil, err
	}
	result.TargetHttpProxyCount = targetHttpProxyResult.TargetHttpProxyCount

	// Execute target TCP proxy workflow
	var targetTcpProxyResult targettcpproxy.GCPComputeTargetTcpProxyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targettcpproxy.GCPComputeTargetTcpProxyWorkflow,
		targettcpproxy.GCPComputeTargetTcpProxyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetTcpProxyResult)
	if err != nil {
		logger.Error("Failed to ingest target TCP proxies", "error", err)
		return nil, err
	}
	result.TargetTcpProxyCount = targetTcpProxyResult.TargetTcpProxyCount

	// Execute target SSL proxy workflow
	var targetSslProxyResult targetsslproxy.GCPComputeTargetSslProxyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targetsslproxy.GCPComputeTargetSslProxyWorkflow,
		targetsslproxy.GCPComputeTargetSslProxyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetSslProxyResult)
	if err != nil {
		logger.Error("Failed to ingest target SSL proxies", "error", err)
		return nil, err
	}
	result.TargetSslProxyCount = targetSslProxyResult.TargetSslProxyCount

	// Execute target HTTPS proxy workflow
	var targetHttpsProxyResult targethttpsproxy.GCPComputeTargetHttpsProxyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targethttpsproxy.GCPComputeTargetHttpsProxyWorkflow,
		targethttpsproxy.GCPComputeTargetHttpsProxyWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetHttpsProxyResult)
	if err != nil {
		logger.Error("Failed to ingest target HTTPS proxies", "error", err)
		return nil, err
	}
	result.TargetHttpsProxyCount = targetHttpsProxyResult.TargetHttpsProxyCount

	// Execute URL map workflow
	var urlMapResult urlmap.GCPComputeUrlMapWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, urlmap.GCPComputeUrlMapWorkflow,
		urlmap.GCPComputeUrlMapWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &urlMapResult)
	if err != nil {
		logger.Error("Failed to ingest URL maps", "error", err)
		return nil, err
	}
	result.UrlMapCount = urlMapResult.UrlMapCount

	// Execute target pool workflow
	var targetPoolResult targetpool.GCPComputeTargetPoolWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, targetpool.GCPComputeTargetPoolWorkflow,
		targetpool.GCPComputeTargetPoolWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &targetPoolResult)
	if err != nil {
		logger.Error("Failed to ingest target pools", "error", err)
		return nil, err
	}
	result.TargetPoolCount = targetPoolResult.TargetPoolCount

	// Execute NEG workflow
	var negResult neg.GCPComputeNegWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, neg.GCPComputeNegWorkflow,
		neg.GCPComputeNegWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &negResult)
	if err != nil {
		logger.Error("Failed to ingest NEGs", "error", err)
		return nil, err
	}
	result.NegCount = negResult.NegCount

	// Execute backend service workflow
	var backendServiceResult backendservice.GCPComputeBackendServiceWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, backendservice.GCPComputeBackendServiceWorkflow,
		backendservice.GCPComputeBackendServiceWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &backendServiceResult)
	if err != nil {
		logger.Error("Failed to ingest backend services", "error", err)
		return nil, err
	}
	result.BackendServiceCount = backendServiceResult.BackendServiceCount

	// Execute NEG endpoint workflow (must run after NEG workflow since it queries NEGs from database)
	var negEndpointResult negendpoint.GCPComputeNegEndpointWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, negendpoint.GCPComputeNegEndpointWorkflow,
		negendpoint.GCPComputeNegEndpointWorkflowParams{ProjectID: params.ProjectID}).Get(ctx, &negEndpointResult)
	if err != nil {
		logger.Error("Failed to ingest NEG endpoints", "error", err)
		return nil, err
	}
	result.NegEndpointCount = negEndpointResult.NegEndpointCount

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
	)

	return result, nil
}
