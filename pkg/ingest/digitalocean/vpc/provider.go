package vpc

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
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
