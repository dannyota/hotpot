package meec

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "meec",
		TaskQueue:          "hotpot-ingest-meec",
		Enabled:            (*config.Service).MEECEnabled,
		RateLimitPerMinute: (*config.Service).MEECRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: MEECInventoryWorkflow,
	})
}
