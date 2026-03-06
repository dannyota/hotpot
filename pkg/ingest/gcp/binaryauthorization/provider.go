package binaryauthorization

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "binaryauthorization",
		Scope:     ingest.ScopeRegional,
		APIName:   "binaryauthorization.googleapis.com",
		Register:  Register,
		Workflow:  GCPBinaryAuthorizationWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPBinaryAuthorizationWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPBinaryAuthorizationWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPBinaryAuthorizationWorkflowResult)
			pr.BinAuthPolicyCount = r.PolicyCount
			pr.AttestorCount = r.AttestorCount
			result.TotalBinAuthPolicies += r.PolicyCount
			result.TotalAttestors += r.AttestorCount
		},
	})
}
