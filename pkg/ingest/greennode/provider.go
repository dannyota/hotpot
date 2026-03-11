package greennode

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "greennode",
		TaskQueue:          "hotpot-ingest-greennode",
		Enabled:            (*config.Service).GreenNodeEnabled,
		RateLimitPerMinute: (*config.Service).GreenNodeRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: GreenNodeInventoryWorkflow,
	})
}
