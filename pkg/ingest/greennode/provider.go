package greennode

import (
	"io"

	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "greennode",
		TaskQueue:          "hotpot-ingest-greennode",
		Enabled:            (*config.Service).GreenNodeEnabled,
		RateLimitPerMinute: (*config.Service).GreenNodeRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, ec *ent.Client) io.Closer {
			return Register(w, cs, ec)
		},
		Workflow: GreenNodeInventoryWorkflow,
	})
}
