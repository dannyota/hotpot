package vault

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
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
