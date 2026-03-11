package computer

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/meec"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "meec",
		Name:      "computer",
		Register:  Register,
		Workflow:  MEECComputerWorkflow,
		NewResult: func() any { return &MEECComputerWorkflowResult{} },
		Aggregate: func(parent *meec.MEECInventoryWorkflowResult, child any) {
			r := child.(*MEECComputerWorkflowResult)
			parent.ComputerCount = r.ComputerCount
		},
	})
}
