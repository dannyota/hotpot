package orgpolicy

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "orgpolicy",
		Scope:     ingest.ScopeGlobal,
		Register:  Register,
		Workflow:  GCPOrgPolicyWorkflow,
		NewParams: func(_, _ string) any { return GCPOrgPolicyWorkflowParams{} },
		NewResult: func() any { return &GCPOrgPolicyWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, _ *gcp.ProjectResult, child any) {
			r := child.(*GCPOrgPolicyWorkflowResult)
			result.TotalConstraints = r.ConstraintCount
			result.TotalOrgPolicies = r.PolicyCount
		},
	})
}
