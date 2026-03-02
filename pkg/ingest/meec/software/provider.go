package software

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/meec"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "meec",
		Name:      "software",
		Register:  Register,
		Workflow:  MEECSoftwareWorkflow,
		NewResult: func() any { return &MEECSoftwareWorkflowResult{} },
		Aggregate: func(parent *meec.MEECInventoryWorkflowResult, child any) {
			r := child.(*MEECSoftwareWorkflowResult)
			parent.SoftwareCount = r.SoftwareCount
		},
	})
}
