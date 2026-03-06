package dns

import (
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest/greennode/dns/hostedzone"
)

// GreenNodeDNSWorkflowParams contains parameters for the DNS workflow.
type GreenNodeDNSWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeDNSWorkflowResult contains the result of the DNS workflow.
type GreenNodeDNSWorkflowResult struct {
	HostedZoneCount int
}

// GreenNodeDNSWorkflow orchestrates GreenNode DNS ingestion.
func GreenNodeDNSWorkflow(ctx workflow.Context, params GreenNodeDNSWorkflowParams) (*GreenNodeDNSWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeDNSWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeDNSWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Hosted Zones
	var hzResult hostedzone.GreenNodeDNSHostedZoneWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, hostedzone.GreenNodeDNSHostedZoneWorkflow, hostedzone.GreenNodeDNSHostedZoneWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &hzResult)
	if err != nil {
		logger.Error("Failed to ingest hosted zones", "error", err)
	} else {
		result.HostedZoneCount = hzResult.HostedZoneCount
	}

	logger.Info("Completed GreenNodeDNSWorkflow",
		"hostedZoneCount", result.HostedZoneCount,
	)

	return result, nil
}
