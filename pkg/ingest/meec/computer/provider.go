package computer

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/meec"
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
