package cloudasset

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "cloudasset",
		Scope:     ingest.ScopeGlobal,
		Register:  Register,
		Workflow:  GCPCloudAssetWorkflow,
		NewParams: func(_, _ string) any { return GCPCloudAssetWorkflowParams{} },
		NewResult: func() any { return &GCPCloudAssetWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, _ *gcp.ProjectResult, child any) {
			r := child.(*GCPCloudAssetWorkflowResult)
			result.TotalAssets = r.AssetCount
			result.TotalAssetPolicies = r.PolicyCount
			result.TotalAssetResources = r.ResourceCount
		},
	})
}
