package dnspolicy

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entdns "github.com/dannyota/hotpot/pkg/storage/ent/gcp/dns"
)

// Register registers DNS policy workflows and activities with the Temporal worker.
// Client is created per activity invocation.
func Register(w worker.Worker, configService *config.Service, entClient *entdns.Client, limiter ratelimit.Limiter) {
	// Create activities with dependencies
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestDNSPolicies)
	w.RegisterWorkflow(GCPDNSPolicyWorkflow)
}
