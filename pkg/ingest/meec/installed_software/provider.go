package installed_software

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/meec"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "meec",
		Name:      "installed_software",
		Register:  Register,
		Workflow:  MEECInstalledSoftwareWorkflow,
		NewResult: func() any { return &MEECInstalledSoftwareWorkflowResult{} },
		Aggregate: func(parent *meec.MEECInventoryWorkflowResult, child any) {
			r := child.(*MEECInstalledSoftwareWorkflowResult)
			parent.InstalledSoftwareCount = r.InstalledSoftwareCount
		},
	})
}
