package endpoint_app

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "endpoint_app",
		Register:  Register,
		Workflow:  S1EndpointAppWorkflow,
		NewResult: func() any { return &S1EndpointAppWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1EndpointAppWorkflowResult)
			parent.EndpointAppCount = r.AppCount
		},
	})
}
