package vpcaccess

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "vpcaccess",
		Scope:     ingest.ScopeRegional,
		APIName:   "vpcaccess.googleapis.com",
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
