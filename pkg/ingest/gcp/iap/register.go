package iap

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/iap/iampolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/iap/settings"
	"entgo.io/ent/dialect"
	entiap "danny.vn/hotpot/pkg/storage/ent/gcp/iap"
)

// Register registers all Identity-Aware Proxy activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entiap.NewClient(entiap.Driver(driver), entiap.AlternateSchema(entiap.DefaultSchemaConfig()))
	settings.Register(w, configService, entClient, limiter)
	iampolicy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAPWorkflow)
}
