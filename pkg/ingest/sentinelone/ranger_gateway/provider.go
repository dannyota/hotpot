package ranger_gateway

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "ranger_gateway",
		Register:  Register,
		Workflow:  S1RangerGatewayWorkflow,
		NewResult: func() any { return &S1RangerGatewayWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1RangerGatewayWorkflowResult)
			parent.RangerGatewayCount = r.GatewayCount
		},
	})
}
