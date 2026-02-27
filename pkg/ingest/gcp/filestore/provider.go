package filestore

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "filestore",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPFilestoreWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPFilestoreWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPFilestoreWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPFilestoreWorkflowResult)
			pr.FilestoreInstanceCount = r.InstanceCount
			result.TotalFilestoreInstances += r.InstanceCount
		},
	})
}
