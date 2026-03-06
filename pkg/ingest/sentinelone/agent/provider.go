package agent

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/sentinelone"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "sentinelone",
		Name:      "agent",
		Register:  Register,
		Workflow:  S1AgentWorkflow,
		NewResult: func() any { return &S1AgentWorkflowResult{} },
		Aggregate: func(parent *sentinelone.S1InventoryWorkflowResult, child any) {
			r := child.(*S1AgentWorkflowResult)
			parent.AgentCount = r.AgentCount
		},
	})
}
