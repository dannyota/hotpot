package container

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "container",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPContainerWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPContainerWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPContainerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPContainerWorkflowResult)
			pr.ClusterCount = r.ClusterCount
			result.TotalClusters += r.ClusterCount
		},
	})
}
