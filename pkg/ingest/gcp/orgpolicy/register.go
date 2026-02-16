package orgpolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/constraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/customconstraint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy/policy"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Organization Policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	constraint.Register(w, configService, entClient, limiter)
	customconstraint.Register(w, configService, entClient, limiter)
	policy.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPOrgPolicyWorkflow)
}
