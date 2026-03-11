package vpc

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "vpc",
		Register:  Register,
		Workflow:  DOVpcWorkflow,
		NewResult: func() any { return &DOVpcWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOVpcWorkflowResult)
			result.VpcCount = r.VpcCount
		},
	})
}
