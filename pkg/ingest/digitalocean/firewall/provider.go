package firewall

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
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
