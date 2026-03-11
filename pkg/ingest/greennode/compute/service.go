package compute

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "compute",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GreenNodeComputeWorkflow,
		NewParams: func(projectID, region, _ string) any {
			return GreenNodeComputeWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeComputeWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeComputeWorkflowResult)
			parent.ServerCount += r.ServerCount
			parent.SSHKeyCount += r.SSHKeyCount
			parent.ServerGroupCount += r.ServerGroupCount
			parent.OSImageCount += r.OSImageCount
			parent.UserImageCount += r.UserImageCount
		},
	})
}
