package rpm

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/reference"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "reference",
		Name:      "rpm",
		Register:  Register,
		Workflow:  RPMPackagesWorkflow,
		NewResult: func() any { return &RPMPackagesWorkflowResult{} },
		Aggregate: func(parent *reference.ReferenceInventoryWorkflowResult, child any) {
			r := child.(*RPMPackagesWorkflowResult)
			parent.RPMPackageCount = r.PackageCount
		},
	})
}
