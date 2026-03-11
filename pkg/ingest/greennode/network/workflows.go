package network

import (
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest/greennode/network/endpoint"
	"danny.vn/hotpot/pkg/ingest/greennode/network/interconnect"
	"danny.vn/hotpot/pkg/ingest/greennode/network/peering"
	"danny.vn/hotpot/pkg/ingest/greennode/network/routetable"
	"danny.vn/hotpot/pkg/ingest/greennode/network/secgroup"
	"danny.vn/hotpot/pkg/ingest/greennode/network/subnet"
	"danny.vn/hotpot/pkg/ingest/greennode/network/vpc"
)

// GreenNodeNetworkWorkflowParams contains parameters for the network workflow.
type GreenNodeNetworkWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkWorkflowResult contains the result of the network workflow.
type GreenNodeNetworkWorkflowResult struct {
	SecgroupCount     int
	EndpointCount     int
	VPCCount          int
	SubnetCount       int
	RouteTableCount   int
	PeeringCount      int
	InterconnectCount int
}

// GreenNodeNetworkWorkflow orchestrates GreenNode network ingestion.
func GreenNodeNetworkWorkflow(ctx workflow.Context, params GreenNodeNetworkWorkflowParams) (*GreenNodeNetworkWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeNetworkWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeNetworkWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Security Groups
	var sgResult secgroup.GreenNodeNetworkSecgroupWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, secgroup.GreenNodeNetworkSecgroupWorkflow, secgroup.GreenNodeNetworkSecgroupWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &sgResult)
	if err != nil {
		logger.Error("Failed to ingest secgroups", "error", err)
	} else {
		result.SecgroupCount = sgResult.SecgroupCount
	}

	// Endpoints
	var epResult endpoint.GreenNodeNetworkEndpointWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, endpoint.GreenNodeNetworkEndpointWorkflow, endpoint.GreenNodeNetworkEndpointWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &epResult)
	if err != nil {
		logger.Error("Failed to ingest endpoints", "error", err)
	} else {
		result.EndpointCount = epResult.EndpointCount
	}

	// VPCs
	var vpcResult vpc.GreenNodeNetworkVPCWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, vpc.GreenNodeNetworkVPCWorkflow, vpc.GreenNodeNetworkVPCWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &vpcResult)
	if err != nil {
		logger.Error("Failed to ingest VPCs", "error", err)
	} else {
		result.VPCCount = vpcResult.VPCCount
	}

	// Subnets
	var subnetResult subnet.GreenNodeNetworkSubnetWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, subnet.GreenNodeNetworkSubnetWorkflow, subnet.GreenNodeNetworkSubnetWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &subnetResult)
	if err != nil {
		logger.Error("Failed to ingest subnets", "error", err)
	} else {
		result.SubnetCount = subnetResult.SubnetCount
	}

	// Route Tables
	var rtResult routetable.GreenNodeNetworkRouteTableWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, routetable.GreenNodeNetworkRouteTableWorkflow, routetable.GreenNodeNetworkRouteTableWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &rtResult)
	if err != nil {
		logger.Error("Failed to ingest route tables", "error", err)
	} else {
		result.RouteTableCount = rtResult.RouteTableCount
	}

	// Peerings
	var peerResult peering.GreenNodeNetworkPeeringWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, peering.GreenNodeNetworkPeeringWorkflow, peering.GreenNodeNetworkPeeringWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &peerResult)
	if err != nil {
		logger.Error("Failed to ingest peerings", "error", err)
	} else {
		result.PeeringCount = peerResult.PeeringCount
	}

	// Interconnects
	var icResult interconnect.GreenNodeNetworkInterconnectWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, interconnect.GreenNodeNetworkInterconnectWorkflow, interconnect.GreenNodeNetworkInterconnectWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &icResult)
	if err != nil {
		logger.Error("Failed to ingest interconnects", "error", err)
	} else {
		result.InterconnectCount = icResult.InterconnectCount
	}

	logger.Info("Completed GreenNodeNetworkWorkflow",
		"secgroupCount", result.SecgroupCount,
		"endpointCount", result.EndpointCount,
		"vpcCount", result.VPCCount,
		"subnetCount", result.SubnetCount,
		"routeTableCount", result.RouteTableCount,
		"peeringCount", result.PeeringCount,
		"interconnectCount", result.InterconnectCount,
	)

	return result, nil
}
