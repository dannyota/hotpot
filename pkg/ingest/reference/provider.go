package reference

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "reference",
		TaskQueue:          "hotpot-ingest-reference",
		Enabled:            (*config.Service).ReferenceEnabled,
		RateLimitPerMinute: (*config.Service).ReferenceRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: ReferenceInventoryWorkflow,
	})
}
