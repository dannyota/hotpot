package sql

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "sql",
		Scope:     ingest.ScopeRegional,
		APIName:   "sqladmin.googleapis.com",
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
