package bigtable

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "bigtable",
		Scope:     ingest.ScopeRegional,
		APIName:   "bigtableadmin.googleapis.com",
		Register:  Register,
		Workflow:  GCPBigtableWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPBigtableWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPBigtableWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPBigtableWorkflowResult)
			pr.BigtableInstanceCount = r.InstanceCount
			pr.BigtableClusterCount = r.ClusterCount
			result.TotalBigtableInstances += r.InstanceCount
			result.TotalBigtableClusters += r.ClusterCount
		},
	})
}
