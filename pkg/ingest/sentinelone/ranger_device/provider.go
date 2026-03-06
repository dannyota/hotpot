package ranger_device

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "ranger_device",
		Register:  Register,
		Workflow:  S1RangerDeviceWorkflow,
		NewResult: func() any { return &S1RangerDeviceWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1RangerDeviceWorkflowResult)
			parent.RangerDeviceCount = r.DeviceCount
		},
	})
}
