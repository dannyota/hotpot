package gcp

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "gcp",
		TaskQueue:          "hotpot-ingest-gcp",
		Enabled:            (*config.Service).GCPEnabled,
		RateLimitPerMinute: (*config.Service).GCPRateLimitPerMinute,
		RegisterWithDriver: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow:     GCPInventoryWorkflow,
		WorkflowArgs: []interface{}{GCPInventoryWorkflowParams{}},
	})
}
