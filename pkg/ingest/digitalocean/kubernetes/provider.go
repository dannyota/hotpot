package kubernetes

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/digitalocean"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "digitalocean",
		Name:      "kubernetes",
		Register:  Register,
		Workflow:  DOKubernetesWorkflow,
		NewResult: func() any { return &DOKubernetesWorkflowResult{} },
		Aggregate: func(result *digitalocean.DOInventoryWorkflowResult, child any) {
			r := child.(*DOKubernetesWorkflowResult)
			result.KubernetesClusterCount = r.ClusterCount
			result.KubernetesNodePoolCount = r.NodePoolCount
		},
	})
}
