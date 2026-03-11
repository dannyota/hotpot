package glb

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "glb",
		Scope:    ingest.ScopeGlobal,
		Register: Register,
		Workflow: GreenNodeGLBWorkflow,
		NewParams: func(projectID, region, _ string) any {
			return GreenNodeGLBWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeGLBWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeGLBWorkflowResult)
			parent.GLBCount = r.GLBCount
			parent.GLBPackageCount = r.PackageCount
			parent.GLBRegionCount = r.RegionCount
		},
	})
}
