package serviceusage

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "serviceusage",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPServiceUsageWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPServiceUsageWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPServiceUsageWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPServiceUsageWorkflowResult)
			pr.EnabledServiceCount = r.ServiceCount
			result.TotalEnabledServices += r.ServiceCount
		},
	})
}
