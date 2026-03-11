package container

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "container",
		Scope:     ingest.ScopeRegional,
		APIName:   "container.googleapis.com",
		Register:  Register,
		Workflow:  GCPContainerWorkflow,
		NewParams: func(projectID, _, quotaProjectID string) any {
			return GCPContainerWorkflowParams{ProjectID: projectID, QuotaProjectID: quotaProjectID}
		},
		NewResult: func() any { return &GCPContainerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPContainerWorkflowResult)
			pr.ClusterCount = r.ClusterCount
			result.TotalClusters += r.ClusterCount
		},
	})
}
