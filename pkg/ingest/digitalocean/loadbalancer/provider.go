package loadbalancer

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "loadbalancer",
		Register:  Register,
		Workflow:  DOLoadBalancerWorkflow,
		NewResult: func() any { return &DOLoadBalancerWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOLoadBalancerWorkflowResult)
			result.LoadBalancerCount = r.LoadBalancerCount
		},
	})
}
