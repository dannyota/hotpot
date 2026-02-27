package iam

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "iam",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPIAMWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPIAMWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPIAMWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPIAMWorkflowResult)
			pr.ServiceAccountCount = r.ServiceAccountCount
			result.TotalServiceAccounts += r.ServiceAccountCount
		},
	})
}
