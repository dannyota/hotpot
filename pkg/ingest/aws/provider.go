package aws

import (
	"io"

	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/ingest"
)

func init() {
	ingest.RegisterProvider(ingest.ProviderRegistration{
		Name:               "aws",
		TaskQueue:          "hotpot-ingest-aws",
		Enabled:            (*config.Service).AWSEnabled,
		RateLimitPerMinute: (*config.Service).AWSRateLimitPerMinute,
		RegisterWithDriver: func(w worker.Worker, cs *config.Service, drv dialect.Driver) io.Closer {
			return Register(w, cs, drv)
		},
		Workflow:     AWSInventoryWorkflow,
		WorkflowArgs: []interface{}{AWSInventoryWorkflowParams{}},
	})
}
