package loadbalancer

import (
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/ingest/greennode"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "greennode",
		Name:     "loadbalancer",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GreenNodeLoadBalancerWorkflow,
		NewParams: func(projectID, region string) any {
			return GreenNodeLoadBalancerWorkflowParams{ProjectID: projectID, Region: region}
		},
		NewResult: func() any { return &GreenNodeLoadBalancerWorkflowResult{} },
		Aggregate: func(parent *greennode.GreenNodeInventoryWorkflowResult, child any) {
			r := child.(*GreenNodeLoadBalancerWorkflowResult)
			parent.LBCount += r.LBCount
			parent.CertificateCount += r.CertificateCount
			parent.LBPackageCount += r.PackageCount
		},
	})
}
