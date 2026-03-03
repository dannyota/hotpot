package alloydb

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "alloydb",
		Scope:     ingest.ScopeRegional,
		APIName:   "alloydb.googleapis.com",
		Register:  Register,
		Workflow:  GCPAlloyDBWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPAlloyDBWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPAlloyDBWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPAlloyDBWorkflowResult)
			pr.AlloyDBClusterCount = r.ClusterCount
			result.TotalAlloyDBClusters += r.ClusterCount
		},
	})
}
