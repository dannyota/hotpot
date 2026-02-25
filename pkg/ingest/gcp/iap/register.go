package iap

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/iampolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap/settings"
	"entgo.io/ent/dialect"
	entiap "github.com/dannyota/hotpot/pkg/storage/ent/gcp/iap"
)

// Register registers all Identity-Aware Proxy activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entiap.NewClient(entiap.Driver(driver), entiap.AlternateSchema(entiap.DefaultSchemaConfig()))
	settings.Register(w, configService, entClient, limiter)
	iampolicy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAPWorkflow)
}
