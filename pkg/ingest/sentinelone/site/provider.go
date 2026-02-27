package site

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "site",
		Register:  Register,
		Workflow:  S1SiteWorkflow,
		NewResult: func() any { return &S1SiteWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1SiteWorkflowResult)
			parent.SiteCount = r.SiteCount
		},
	})
}
