package accesscontextmanager

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "accesscontextmanager",
		Scope:     ingest.ScopeGlobal,
		Register:  Register,
		Workflow:  GCPAccessContextManagerWorkflow,
		NewParams: func(_, _ string) any { return GCPAccessContextManagerWorkflowParams{} },
		NewResult: func() any { return &GCPAccessContextManagerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, _ *gcp.ProjectResult, child any) {
			r := child.(*GCPAccessContextManagerWorkflowResult)
			result.TotalAccessPolicies = r.PolicyCount
			result.TotalAccessLevels = r.LevelCount
			result.TotalServicePerimeters = r.PerimeterCount
		},
	})
}
