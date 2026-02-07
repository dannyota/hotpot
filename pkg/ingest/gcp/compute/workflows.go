package compute

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"hotpot/pkg/ingest/gcp/compute/address"
	"hotpot/pkg/ingest/gcp/compute/disk"
	"hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"hotpot/pkg/ingest/gcp/compute/globaladdress"
	"hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"hotpot/pkg/ingest/gcp/compute/healthcheck"
	"hotpot/pkg/ingest/gcp/compute/image"
	"hotpot/pkg/ingest/gcp/compute/instance"
	"hotpot/pkg/ingest/gcp/compute/instancegroup"
	"hotpot/pkg/ingest/gcp/compute/network"
	"hotpot/pkg/ingest/gcp/compute/snapshot"
	"hotpot/pkg/ingest/gcp/compute/subnetwork"
	"hotpot/pkg/ingest/gcp/compute/targetinstance"
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
	)

	return result, nil
}
