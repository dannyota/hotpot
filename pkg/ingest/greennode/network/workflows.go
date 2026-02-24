package network

import (
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/endpoint"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/secgroup"
)

// GreenNodeNetworkWorkflowParams contains parameters for the network workflow.
type GreenNodeNetworkWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeNetworkWorkflowResult contains the result of the network workflow.
type GreenNodeNetworkWorkflowResult struct {
	SecgroupCount int
	EndpointCount int
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

	logger.Info("Completed GreenNodeNetworkWorkflow",
		"secgroupCount", result.SecgroupCount,
		"endpointCount", result.EndpointCount,
	)

	return result, nil
}
