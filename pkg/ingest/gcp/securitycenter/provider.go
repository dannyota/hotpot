package securitycenter

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "securitycenter",
		Scope:     ingest.ScopeGlobal,
		Register:  Register,
		Workflow:  GCPSecurityCenterWorkflow,
		NewParams: func(_, _ string) any { return GCPSecurityCenterWorkflowParams{} },
		NewResult: func() any { return &GCPSecurityCenterWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, _ *gcp.ProjectResult, child any) {
			r := child.(*GCPSecurityCenterWorkflowResult)
			result.TotalSources = r.SourceCount
			result.TotalFindings = r.FindingCount
		},
	})
}
