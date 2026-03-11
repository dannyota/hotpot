package apicatalog

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "apicatalog",
		TaskQueue:          "hotpot-ingest-apicatalog",
		Enabled:            (*config.Service).ApiCatalogEnabled,
		RateLimitPerMinute: func(*config.Service) int { return 120 },
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			Register(w, cs, drv)
			return nil
		},
		Workflow:     ApiCatalogWorkflow,
		WorkflowArgs: []interface{}{ApiCatalogWorkflowParams{}},
	})
}
