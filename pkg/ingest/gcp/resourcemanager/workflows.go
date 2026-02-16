package resourcemanager

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/folder"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/folderiampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/orgiampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/organization"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/project"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager/projectiampolicy"
)

// GCPResourceManagerWorkflowParams contains parameters for the resource manager workflow.
type GCPResourceManagerWorkflowParams struct {
	// Empty - discovers all accessible resources
}

// GCPResourceManagerWorkflowResult contains the result of the resource manager workflow.
type GCPResourceManagerWorkflowResult struct {
	ProjectCount          int
	ProjectIDs            []string
	OrganizationCount     int
	FolderCount           int
	OrgIamPolicyCount     int
	FolderIamPolicyCount  int
	ProjectIamPolicyCount int
	DurationMillis        int64
}

// GCPResourceManagerWorkflow ingests all GCP Resource Manager resources.
// Orchestrates child workflows - each manages its own session and client lifecycle.
func GCPResourceManagerWorkflow(ctx workflow.Context, params GCPResourceManagerWorkflowParams) (*GCPResourceManagerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPResourceManagerWorkflow")

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

	result := &GCPResourceManagerWorkflowResult{}
	startTime := workflow.Now(ctx)

	// Phase 1: Discover parent resources (project, organization, folder) - independent of each other

	// Execute project workflow
	var projectResult project.GCPResourceManagerProjectWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, project.GCPResourceManagerProjectWorkflow,
		project.GCPResourceManagerProjectWorkflowParams{}).Get(ctx, &projectResult)
	if err != nil {
		logger.Error("Failed to ingest projects", "error", err)
		return nil, err
	}
	result.ProjectCount = projectResult.ProjectCount
	result.ProjectIDs = projectResult.ProjectIDs

	// Execute organization workflow
	var orgResult organization.GCPResourceManagerOrganizationWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, organization.GCPResourceManagerOrganizationWorkflow,
		organization.GCPResourceManagerOrganizationWorkflowParams{}).Get(ctx, &orgResult)
	if err != nil {
		logger.Error("Failed to ingest organizations", "error", err)
		return nil, err
	}
	result.OrganizationCount = orgResult.OrganizationCount

	// Execute folder workflow
	var folderResult folder.GCPResourceManagerFolderWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, folder.GCPResourceManagerFolderWorkflow,
		folder.GCPResourceManagerFolderWorkflowParams{}).Get(ctx, &folderResult)
	if err != nil {
		logger.Error("Failed to ingest folders", "error", err)
		return nil, err
	}
	result.FolderCount = folderResult.FolderCount

	// Phase 2: IAM policies (must run after parent resources since they query from database)

	// Execute org IAM policy workflow (queries organizations from DB)
	var orgIamResult orgiampolicy.GCPResourceManagerOrgIamPolicyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, orgiampolicy.GCPResourceManagerOrgIamPolicyWorkflow,
		orgiampolicy.GCPResourceManagerOrgIamPolicyWorkflowParams{}).Get(ctx, &orgIamResult)
	if err != nil {
		logger.Error("Failed to ingest org IAM policies", "error", err)
		return nil, err
	}
	result.OrgIamPolicyCount = orgIamResult.PolicyCount

	// Execute folder IAM policy workflow (queries folders from DB)
	var folderIamResult folderiampolicy.GCPResourceManagerFolderIamPolicyWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, folderiampolicy.GCPResourceManagerFolderIamPolicyWorkflow,
		folderiampolicy.GCPResourceManagerFolderIamPolicyWorkflowParams{}).Get(ctx, &folderIamResult)
	if err != nil {
		logger.Error("Failed to ingest folder IAM policies", "error", err)
		return nil, err
	}
	result.FolderIamPolicyCount = folderIamResult.PolicyCount

	// Execute project IAM policy workflows (one per discovered project, queries projects from DB)
	for _, pid := range result.ProjectIDs {
		var projectIamResult projectiampolicy.GCPResourceManagerProjectIamPolicyWorkflowResult
		err = workflow.ExecuteChildWorkflow(childCtx, projectiampolicy.GCPResourceManagerProjectIamPolicyWorkflow,
			projectiampolicy.GCPResourceManagerProjectIamPolicyWorkflowParams{ProjectID: pid}).Get(ctx, &projectIamResult)
		if err != nil {
			logger.Error("Failed to ingest project IAM policy", "projectID", pid, "error", err)
			return nil, err
		}
		result.ProjectIamPolicyCount += projectIamResult.PolicyCount
	}

	result.DurationMillis = workflow.Now(ctx).Sub(startTime).Milliseconds()

	logger.Info("Completed GCPResourceManagerWorkflow",
		"projectCount", result.ProjectCount,
		"organizationCount", result.OrganizationCount,
		"folderCount", result.FolderCount,
		"orgIamPolicyCount", result.OrgIamPolicyCount,
		"folderIamPolicyCount", result.FolderIamPolicyCount,
		"projectIamPolicyCount", result.ProjectIamPolicyCount,
	)

	return result, nil
}
