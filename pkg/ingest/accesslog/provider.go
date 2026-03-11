package accesslog

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "accesslog",
		TaskQueue:          "hotpot-ingest-accesslog",
		Enabled:            (*config.Service).AccessLogEnabled,
		RateLimitPerMinute: (*config.Service).AccessLogRateLimitPerMinute,
		Register: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			Register(w, cs, drv)
			return nil
		},
		Workflow: AccessLogWorkflow,
	})
}
