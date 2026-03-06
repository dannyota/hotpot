package key

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "key",
		Register:  Register,
		Workflow:  DOKeyWorkflow,
		NewResult: func() any { return &DOKeyWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOKeyWorkflowResult)
			result.KeyCount = r.KeyCount
		},
	})
}
