package volume

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "volume",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GreenNodeVolumeWorkflow,
		NewParams: func(projectID, region string) any {
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
