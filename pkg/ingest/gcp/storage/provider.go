package storage

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "storage",
		Scope:     ingest.ScopeRegional,
		APIName:   "storage.googleapis.com",
		Register:  Register,
		Workflow:  GCPStorageWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPStorageWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPStorageWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPStorageWorkflowResult)
			pr.BucketCount = r.BucketCount
			pr.BucketIamPolicyCount = r.BucketIamPolicyCount
			result.TotalBuckets += r.BucketCount
			result.TotalBucketIamPolicies += r.BucketIamPolicyCount
		},
	})
}
