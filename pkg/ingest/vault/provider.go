package vault

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "vault",
		TaskQueue:          "hotpot-ingest-vault",
		Enabled:            (*config.Service).VaultEnabled,
		RateLimitPerMinute: (*config.Service).VaultRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow: VaultInventoryWorkflow,
	})
}
