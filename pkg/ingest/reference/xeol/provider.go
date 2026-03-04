package xeol

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/reference"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "reference",
		Name:      "xeol",
		Register:  Register,
		Workflow:  XeolWorkflow,
		NewResult: func() any { return &XeolWorkflowResult{} },
		Aggregate: func(parent *reference.ReferenceInventoryWorkflowResult, child any) {
			r := child.(*XeolWorkflowResult)
			parent.XeolProductCount = r.ProductCount
			parent.XeolCycleCount = r.CycleCount
			parent.XeolPurlCount = r.PurlCount
			parent.XeolVulnCount = r.VulnCount
		},
	})
}
