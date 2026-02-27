package vpcaccess

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "vpcaccess",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPVpcAccessWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPVpcAccessWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPVpcAccessWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPVpcAccessWorkflowResult)
			pr.ConnectorCount = r.ConnectorCount
			result.TotalConnectors += r.ConnectorCount
		},
	})
}
