package sentinelone

import (
	"io"

	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "s1",
		TaskQueue:          "hotpot-ingest-s1",
		Enabled:            (*config.Service).S1Enabled,
		RateLimitPerMinute: (*config.Service).S1RateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, ec *ent.Client) io.Closer {
			return Register(w, cs, ec)
		},
	})
}
