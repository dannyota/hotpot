package volume

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "volume",
		Register:  Register,
		Workflow:  DOVolumeWorkflow,
		NewResult: func() any { return &DOVolumeWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOVolumeWorkflowResult)
			result.VolumeCount = r.VolumeCount
		},
	})
}
