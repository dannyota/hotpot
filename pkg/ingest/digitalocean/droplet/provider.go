package droplet

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "droplet",
		Register:  Register,
		Workflow:  DODropletWorkflow,
		NewResult: func() any { return &DODropletWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DODropletWorkflowResult)
			result.DropletCount = r.DropletCount
		},
	})
}
