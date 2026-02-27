package redis

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "redis",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  GCPRedisWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPRedisWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPRedisWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPRedisWorkflowResult)
			pr.RedisInstanceCount = r.InstanceCount
			result.TotalRedisInstances += r.InstanceCount
		},
	})
}
