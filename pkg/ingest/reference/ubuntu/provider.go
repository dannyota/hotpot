package ubuntu

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/reference"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "reference",
		Name:      "ubuntu",
		Register:  Register,
		Workflow:  UbuntuPackagesWorkflow,
		NewResult: func() any { return &UbuntuPackagesWorkflowResult{} },
		Aggregate: func(parent *reference.ReferenceInventoryWorkflowResult, child any) {
			r := child.(*UbuntuPackagesWorkflowResult)
			parent.UbuntuPackageCount = r.PackageCount
		},
	})
}
