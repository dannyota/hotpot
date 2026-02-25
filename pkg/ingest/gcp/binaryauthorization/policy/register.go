package policy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entbinaryauthorization "github.com/dannyota/hotpot/pkg/storage/ent/gcp/binaryauthorization"
)

// Register registers all Binary Authorization policy activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entbinaryauthorization.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestBinaryAuthorizationPolicies)
	w.RegisterWorkflow(GCPBinaryAuthorizationPolicyWorkflow)
}
