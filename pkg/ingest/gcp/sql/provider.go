package sql

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "sql",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPSQLWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPSQLWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPSQLWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPSQLWorkflowResult)
			pr.SQLInstanceCount = r.InstanceCount
			result.TotalSQLInstances += r.InstanceCount
		},
	})
}
