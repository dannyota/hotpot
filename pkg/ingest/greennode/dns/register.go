package dns

import (
	"entgo.io/ent/dialect"
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/dns/hostedzone"
	entdns "danny.vn/hotpot/pkg/storage/ent/greennode/dns"
)

// Register registers all GreenNode DNS activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entdns.NewClient(entdns.Driver(driver), entdns.AlternateSchema(entdns.DefaultSchemaConfig()))

	hostedzone.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeDNSWorkflow)
}
