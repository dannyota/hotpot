package digitalocean

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "do",
		TaskQueue:          "hotpot-ingest-do",
		Enabled:            (*config.Service).DOEnabled,
		RateLimitPerMinute: (*config.Service).DORateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: DOInventoryWorkflow,
	})
}
