package cloudfunctions

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "cloudfunctions",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPCloudFunctionsWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPCloudFunctionsWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPCloudFunctionsWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPCloudFunctionsWorkflowResult)
			pr.FunctionCount = r.FunctionCount
			result.TotalFunctions += r.FunctionCount
		},
	})
}
