package group

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "group",
		Register:  Register,
		Workflow:  S1GroupWorkflow,
		NewResult: func() any { return &S1GroupWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1GroupWorkflowResult)
			parent.GroupCount = r.GroupCount
		},
	})
}
