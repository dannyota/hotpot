package dns

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns/dnspolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns/managedzone"
	"entgo.io/ent/dialect"
	entdns "github.com/dannyota/hotpot/pkg/storage/ent/gcp/dns"
)

// Register registers all DNS activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entdns.NewClient(entdns.Driver(driver), entdns.AlternateSchema(entdns.DefaultSchemaConfig()))
	managedzone.Register(w, configService, entClient, limiter)
	dnspolicy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPDNSWorkflow)
}
