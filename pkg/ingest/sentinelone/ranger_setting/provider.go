package ranger_setting

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "ranger_setting",
		Register:  Register,
		Workflow:  S1RangerSettingWorkflow,
		NewResult: func() any { return &S1RangerSettingWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1RangerSettingWorkflowResult)
			parent.RangerSettingCount = r.SettingCount
		},
	})
}
