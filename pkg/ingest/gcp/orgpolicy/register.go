package orgpolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/constraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/customconstraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/policy"
	"entgo.io/ent/dialect"
	entorgpolicy "github.com/dannyota/hotpot/pkg/storage/ent/gcp/orgpolicy"
	entresourcemanager "github.com/dannyota/hotpot/pkg/storage/ent/gcp/resourcemanager"
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
