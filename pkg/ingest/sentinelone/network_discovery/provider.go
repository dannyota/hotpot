package network_discovery

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "network_discovery",
		Register:  Register,
		Workflow:  S1NetworkDiscoveryWorkflow,
		NewResult: func() any { return &S1NetworkDiscoveryWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1NetworkDiscoveryWorkflowResult)
			parent.NetworkDiscoveryCount = r.DeviceCount
		},
	})
}
