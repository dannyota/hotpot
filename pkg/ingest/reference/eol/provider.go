package eol

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/reference"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "reference",
		Name:      "eol",
		Register:  Register,
		Workflow:  EOLWorkflow,
		NewResult: func() any { return &EOLWorkflowResult{} },
		Aggregate: func(parent *reference.ReferenceInventoryWorkflowResult, child any) {
			r := child.(*EOLWorkflowResult)
			parent.EOLProductCount = r.ProductCount
			parent.EOLCycleCount = r.CycleCount
		},
	})
}
