package network

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "network",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GreenNodeNetworkWorkflow,
		NewParams: func(projectID, region, _ string) any {
			return GreenNodeNetworkWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeNetworkWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeNetworkWorkflowResult)
			parent.SecgroupCount += r.SecgroupCount
			parent.EndpointCount += r.EndpointCount
			parent.VPCCount += r.VPCCount
			parent.SubnetCount += r.SubnetCount
			parent.RouteTableCount += r.RouteTableCount
			parent.PeeringCount += r.PeeringCount
			parent.InterconnectCount += r.InterconnectCount
		},
	})
}
