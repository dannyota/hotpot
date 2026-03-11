package iam

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "iam",
		Scope:     ingest.ScopeRegional,
		APIName:   "iam.googleapis.com",
		Register:  Register,
		Workflow:  GCPIAMWorkflow,
		NewParams: func(projectID, _, quotaProjectID string) any {
			return GCPIAMWorkflowParams{ProjectID: projectID, QuotaProjectID: quotaProjectID}
		},
		NewResult: func() any { return &GCPIAMWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPIAMWorkflowResult)
			pr.ServiceAccountCount = r.ServiceAccountCount
			result.TotalServiceAccounts += r.ServiceAccountCount
		},
	})
}
