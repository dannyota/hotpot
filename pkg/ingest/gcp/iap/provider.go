package iap

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "iap",
		Scope:     ingest.ScopeRegional,
		APIName:   "iap.googleapis.com",
		Register:  Register,
		Workflow:  GCPIAPWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPIAPWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPIAPWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPIAPWorkflowResult)
			pr.IAPSettingsCount = r.SettingsCount
			pr.IAPPolicyCount = r.PolicyCount
			result.TotalIAPSettings += r.SettingsCount
			result.TotalIAPPolicies += r.PolicyCount
		},
	})
}
