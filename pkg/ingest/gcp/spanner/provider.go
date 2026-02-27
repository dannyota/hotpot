package spanner

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "spanner",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPSpannerWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPSpannerWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPSpannerWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPSpannerWorkflowResult)
			pr.SpannerInstanceCount = r.InstanceCount
			pr.SpannerDatabaseCount = r.DatabaseCount
			result.TotalSpannerInstances += r.InstanceCount
			result.TotalSpannerDatabases += r.DatabaseCount
		},
	})
}
