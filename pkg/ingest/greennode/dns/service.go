package dns

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "dns",
		Scope:    ingest.ScopeGlobal,
		Register: Register,
		Workflow: GreenNodeDNSWorkflow,
		NewParams: func(projectID, region string) any {
			return GreenNodeDNSWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeDNSWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeDNSWorkflowResult)
			parent.HostedZoneCount = r.HostedZoneCount
		},
	})
}
