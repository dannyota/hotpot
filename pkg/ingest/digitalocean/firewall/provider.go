package firewall

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "firewall",
		Register:  Register,
		Workflow:  DOFirewallWorkflow,
		NewResult: func() any { return &DOFirewallWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOFirewallWorkflowResult)
			result.FirewallCount = r.FirewallCount
		},
	})
}
