package account

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "account",
		Register:  Register,
		Workflow:  S1AccountWorkflow,
		NewResult: func() any { return &S1AccountWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1AccountWorkflowResult)
			parent.AccountCount = r.AccountCount
		},
	})
}
