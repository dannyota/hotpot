package digitalocean

import (
	"io"

	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "do",
		TaskQueue:          "hotpot-ingest-do",
		Enabled:            (*config.Service).DOEnabled,
		RateLimitPerMinute: (*config.Service).DORateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, ec *ent.Client) io.Closer {
			return Register(w, cs, ec)
		},
	})
}
