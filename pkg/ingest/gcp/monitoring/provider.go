package monitoring

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/gcp"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider:  "gcp",
		Name:      "monitoring",
		Scope:     ingest.ScopeRegional,
		APIName:   "monitoring.googleapis.com",
		Register:  Register,
		Workflow:  GCPMonitoringWorkflow,
		NewParams: func(projectID, _ string) any {
			return GCPMonitoringWorkflowParams{ProjectID: projectID}
		},
		NewResult: func() any { return &GCPMonitoringWorkflowResult{} },
		Aggregate: func(result *gcp.GCPInventoryWorkflowResult, pr *gcp.ProjectResult, child any) {
			r := child.(*GCPMonitoringWorkflowResult)
			pr.AlertPolicyCount = r.AlertPolicyCount
			pr.UptimeCheckCount = r.UptimeCheckCount
			result.TotalAlertPolicies += r.AlertPolicyCount
			result.TotalUptimeChecks += r.UptimeCheckCount
		},
	})
}
