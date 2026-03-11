package job

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/jenkins"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "jenkins",
		Name:      "job",
		Register:  Register,
		Workflow:  JenkinsJobWorkflow,
		NewResult: func() any { return &JenkinsJobWorkflowResult{} },
		Aggregate: func(result *jenkins.JenkinsInventoryWorkflowResult, child any) {
			r := child.(*JenkinsJobWorkflowResult)
			result.JobCount = r.JobCount
			result.BuildCount = r.BuildCount
			result.RepoCount = r.RepoCount
		},
	})
}
