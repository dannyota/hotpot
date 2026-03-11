package orgpolicy

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/constraint"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/customconstraint"
	"danny.vn/hotpot/pkg/ingest/gcp/orgpolicy/policy"
	"entgo.io/ent/dialect"
	entorgpolicy "danny.vn/hotpot/pkg/storage/ent/gcp/orgpolicy"
	entresourcemanager "danny.vn/hotpot/pkg/storage/ent/gcp/resourcemanager"
)

// Register registers all Organization Policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entorgpolicy.NewClient(entorgpolicy.Driver(driver), entorgpolicy.AlternateSchema(entorgpolicy.DefaultSchemaConfig()))
	rmClient := entresourcemanager.NewClient(entresourcemanager.Driver(driver), entresourcemanager.AlternateSchema(entresourcemanager.DefaultSchemaConfig()))
	constraint.Register(w, configService, entClient, rmClient, limiter)
	customconstraint.Register(w, configService, entClient, rmClient, limiter)
	policy.Register(w, configService, entClient, rmClient, limiter)

	w.RegisterWorkflow(GCPOrgPolicyWorkflow)
}
