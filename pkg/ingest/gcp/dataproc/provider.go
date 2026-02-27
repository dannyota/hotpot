package dataproc

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "dataproc",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPDataprocWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPDataprocWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPDataprocWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPDataprocWorkflowResult)
			pr.DataprocClusterCount = r.ClusterCount
			result.TotalDataprocClusters += r.ClusterCount
		},
	})
}
