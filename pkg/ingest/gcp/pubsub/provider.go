package pubsub

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "pubsub",
		Scope:     ingest.ScopeRegional,
		APIName:   "pubsub.googleapis.com",
		Register:  Register,
		Workflow:  GCPPubSubWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPPubSubWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPPubSubWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPPubSubWorkflowResult)
			pr.TopicCount = r.TopicCount
			pr.SubscriptionCount = r.SubscriptionCount
			result.TotalTopics += r.TopicCount
			result.TotalSubscriptions += r.SubscriptionCount
		},
	})
}
