package cpe

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/reference"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "reference",
		Name:      "cpe",
		Register:  Register,
		Workflow:  CPEWorkflow,
		NewResult: func() any { return &CPEWorkflowResult{} },
		Aggregate: func(parent *reference.ReferenceInventoryWorkflowResult, child any) {
			r := child.(*CPEWorkflowResult)
			parent.CPECount = r.CPECount
		},
	})
}
