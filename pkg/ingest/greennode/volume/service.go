package volume

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "volume",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GreenNodeVolumeWorkflow,
		NewParams: func(projectID, region, _ string) any {
			return GreenNodeVolumeWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeVolumeWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeVolumeWorkflowResult)
			parent.BlockVolumeCount += r.BlockVolumeCount
			parent.VolumeTypeCount += r.VolumeTypeCount
			parent.VolumeTypeZoneCount += r.VolumeTypeZoneCount
		},
	})
}
