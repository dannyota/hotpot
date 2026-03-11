package gcplogging

import (
	"danny.vn/hotpot/pkg/ingest"
	"danny.vn/hotpot/pkg/ingest/accesslog"
)

func init() {
	ingest.RegisterService(ingest.ServiceRegistration{
		Provider: "accesslog",
		Name:     "gcplogging",
		Scope:    ingest.ScopeRegional,
		Register: Register,
		Workflow: GcpLoggingTrafficWorkflow,
		NewParams: func(_, _, _ string) any {
			return accesslog.ServiceWorkflowParams{}
		},
		NewResult: func() any { return &accesslog.ServiceWorkflowResult{} },
	})
}
