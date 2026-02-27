package ranger_setting

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone"
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
