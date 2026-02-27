package project

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "project",
		Register:  Register,
		Workflow:  DOProjectWorkflow,
		NewResult: func() any { return &DOProjectWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOProjectWorkflowResult)
			result.ProjectCount = r.ProjectCount
			result.ResourceCount = r.ResourceCount
		},
	})
}
