package compute

import (
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest/greennode/compute/osimage"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/server"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/servergroup"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/sshkey"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/userimage"
)

// GreenNodeComputeWorkflowParams contains parameters for the compute workflow.
type GreenNodeComputeWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeComputeWorkflowResult contains the result of the compute workflow.
type GreenNodeComputeWorkflowResult struct {
	ServerCount      int
	SSHKeyCount      int
	ServerGroupCount int
	OSImageCount     int
	UserImageCount   int
}

// GreenNodeComputeWorkflow orchestrates GreenNode compute ingestion.
func GreenNodeComputeWorkflow(ctx workflow.Context, params GreenNodeComputeWorkflowParams) (*GreenNodeComputeWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeComputeWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeComputeWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Servers
	var serverResult server.GreenNodeComputeServerWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, server.GreenNodeComputeServerWorkflow, server.GreenNodeComputeServerWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &serverResult)
	if err != nil {
		logger.Error("Failed to ingest servers", "error", err)
	} else {
		result.ServerCount = serverResult.ServerCount
	}

	// SSH Keys
	var sshKeyResult sshkey.GreenNodeComputeSSHKeyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, sshkey.GreenNodeComputeSSHKeyWorkflow, sshkey.GreenNodeComputeSSHKeyWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &sshKeyResult)
	if err != nil {
		logger.Error("Failed to ingest SSH keys", "error", err)
	} else {
		result.SSHKeyCount = sshKeyResult.KeyCount
	}

	// Server Groups
	var sgResult servergroup.GreenNodeComputeServerGroupWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, servergroup.GreenNodeComputeServerGroupWorkflow, servergroup.GreenNodeComputeServerGroupWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &sgResult)
	if err != nil {
		logger.Error("Failed to ingest server groups", "error", err)
	} else {
		result.ServerGroupCount = sgResult.GroupCount
	}

	// OS Images
	var osImageResult osimage.GreenNodeComputeOSImageWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, osimage.GreenNodeComputeOSImageWorkflow, osimage.GreenNodeComputeOSImageWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &osImageResult)
	if err != nil {
		logger.Error("Failed to ingest OS images", "error", err)
	} else {
		result.OSImageCount = osImageResult.OSImageCount
	}

	// User Images
	var userImageResult userimage.GreenNodeComputeUserImageWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, userimage.GreenNodeComputeUserImageWorkflow, userimage.GreenNodeComputeUserImageWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &userImageResult)
	if err != nil {
		logger.Error("Failed to ingest user images", "error", err)
	} else {
		result.UserImageCount = userImageResult.UserImageCount
	}

	logger.Info("Completed GreenNodeComputeWorkflow",
		"serverCount", result.ServerCount,
		"sshKeyCount", result.SSHKeyCount,
		"serverGroupCount", result.ServerGroupCount,
		"osImageCount", result.OSImageCount,
		"userImageCount", result.UserImageCount,
	)

	return result, nil
}
