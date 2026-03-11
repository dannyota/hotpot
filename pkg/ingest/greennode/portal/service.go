package portal

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "portal",
		Scope:    ingest.ScopeGlobal,
		Register: Register,
		Workflow: GreenNodePortalWorkflow,
		NewParams: func(projectID, region, _ string) any {
			return GreenNodePortalWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodePortalWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodePortalWorkflowResult)
			parent.RegionCount = r.RegionCount
			parent.QuotaCount = r.QuotaCount
			parent.ZoneCount = r.ZoneCount
		},
	})
}
