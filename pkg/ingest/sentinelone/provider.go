package sentinelone

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "s1",
		TaskQueue:          "hotpot-ingest-s1",
		Enabled:            (*config.Service).S1Enabled,
		RateLimitPerMinute: (*config.Service).S1RateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: S1InventoryWorkflow,
	})
}
