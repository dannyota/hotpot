package run

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "run",
		Scope:     ingest.ScopeRegional,
		APIName:   "run.googleapis.com",
		Register:  Register,
		Workflow:  GCPRunWorkflow,
		NewParams: func(projectID, _, _ string) any {
			return GCPRunWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPRunWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPRunWorkflowResult)
			pr.RunServiceCount = r.ServiceCount
			pr.RunRevisionCount = r.RevisionCount
			result.TotalRunServices += r.ServiceCount
			result.TotalRunRevisions += r.RevisionCount
		},
	})
}
