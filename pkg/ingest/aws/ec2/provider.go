package ec2

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/aws"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "aws",
		Name:      "ec2",
		Scope:     ingest.ScopeRegional,
		Register:  Register,
		Workflow:  AWSEC2Workflow,
		NewParams: func(_, region string) any {
			return AWSEC2WorkflowParams{Region: region}
		},
		NewResult: func() any { return &AWSEC2WorkflowResult{} },
		Aggregate: func(result *aws.AWSInventoryWorkflowResult, rr *aws.RegionResult, child any) {
			r := child.(*AWSEC2WorkflowResult)
			rr.InstanceCount = r.InstanceCount
			result.TotalInstances += r.InstanceCount
		},
	})
}
