package greennode

import (
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

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Portal (regions, quotas)
	var portalResult portal.GreenNodePortalWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, portal.GreenNodePortalWorkflow).Get(ctx, &portalResult)
	if err != nil {
		logger.Error("Failed portal ingestion", "error", err)
	} else {
		result.RegionCount = portalResult.RegionCount
		result.QuotaCount = portalResult.QuotaCount
	}

	// Compute (servers, SSH keys, server groups)
	var computeResult compute.GreenNodeComputeWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, compute.GreenNodeComputeWorkflow, compute.GreenNodeComputeWorkflowParams{
		ProjectID: "", // Set by config at activity level
	}).Get(ctx, &computeResult)
	if err != nil {
		logger.Error("Failed compute ingestion", "error", err)
	} else {
		result.ServerCount = computeResult.ServerCount
		result.SSHKeyCount = computeResult.SSHKeyCount
		result.ServerGroupCount = computeResult.ServerGroupCount
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
