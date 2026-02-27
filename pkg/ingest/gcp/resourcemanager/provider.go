package resourcemanager

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "resourcemanager",
		Scope:     ingest.ScopeGlobal,
		Register:  Register,
		Workflow:  GCPResourceManagerWorkflow,
		NewParams: func(_, _ string) any { return GCPResourceManagerWorkflowParams{} },
		NewResult: func() any { return &GCPResourceManagerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, _ *gcp.ProjectResult, child any) {
			r := child.(*GCPResourceManagerWorkflowResult)
			result.TotalProjects = r.ProjectCount
			result.TotalOrganizations = r.OrganizationCount
			result.TotalFolders = r.FolderCount
			result.TotalOrgIamPolicies = r.OrgIamPolicyCount
			result.TotalFolderIamPolicies = r.FolderIamPolicyCount
			result.TotalProjectIamPolicies = r.ProjectIamPolicyCount
		},
	})
}
