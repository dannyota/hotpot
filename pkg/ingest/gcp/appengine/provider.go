package appengine

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "appengine",
		Scope:     ingest.ScopeRegional,
		APIName:   "appengine.googleapis.com",
		Register:  Register,
		Workflow:  GCPAppEngineWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPAppEngineWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPAppEngineWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPAppEngineWorkflowResult)
			pr.ApplicationCount = r.ApplicationCount
			pr.AppServiceCount = r.ServiceCount
			result.TotalApplications += r.ApplicationCount
			result.TotalAppServices += r.ServiceCount
		},
	})
}
