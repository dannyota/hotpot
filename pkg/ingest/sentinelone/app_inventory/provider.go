package app_inventory

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "app_inventory",
		Register:  Register,
		Workflow:  S1AppInventoryWorkflow,
		NewResult: func() any { return &S1AppInventoryWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1AppInventoryWorkflowResult)
			parent.AppInventoryCount = r.AppCount
		},
	})
}
