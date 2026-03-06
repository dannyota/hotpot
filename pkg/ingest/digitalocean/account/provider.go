package account

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "account",
		Register:  Register,
		Workflow:  DOAccountWorkflow,
		NewResult: func() any { return &DOAccountWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOAccountWorkflowResult)
			result.AccountCount = r.AccountCount
		},
	})
}
