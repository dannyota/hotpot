package domain

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "domain",
		Register:  Register,
		Workflow:  DODomainWorkflow,
		NewResult: func() any { return &DODomainWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DODomainWorkflowResult)
			result.DomainCount = r.DomainCount
			result.DomainRecordCount = r.RecordCount
		},
	})
}
