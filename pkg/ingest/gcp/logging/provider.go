package logging

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "logging",
		Scope:     ingest.ScopeRegional,
		APIName:   "logging.googleapis.com",
		Register:  Register,
		Workflow:  GCPLoggingWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPLoggingWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPLoggingWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPLoggingWorkflowResult)
			pr.SinkCount = r.SinkCount
			pr.LogBucketCount = r.BucketCount
			pr.LogMetricCount = r.LogMetricCount
			pr.LogExclusionCount = r.ExclusionCount
			result.TotalSinks += r.SinkCount
			result.TotalLogBuckets += r.BucketCount
			result.TotalLogMetrics += r.LogMetricCount
			result.TotalLogExclusions += r.ExclusionCount
		},
	})
}
