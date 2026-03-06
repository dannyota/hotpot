package xeol

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/reference"
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
