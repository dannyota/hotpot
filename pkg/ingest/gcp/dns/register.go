package dns

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/dns/dnspolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/dns/managedzone"
	"entgo.io/ent/dialect"
	entdns "danny.vn/hotpot/pkg/storage/ent/gcp/dns"
)

// Register registers all DNS activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entdns.NewClient(entdns.Driver(driver), entdns.AlternateSchema(entdns.DefaultSchemaConfig()))
	managedzone.Register(w, configService, entClient, limiter)
	dnspolicy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPDNSWorkflow)
}
