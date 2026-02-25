package subscription

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entpubsub "github.com/dannyota/hotpot/pkg/storage/ent/gcp/pubsub"
)

// Register registers all Pub/Sub subscription activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entpubsub.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestPubSubSubscriptions)
	w.RegisterWorkflow(GCPPubSubSubscriptionWorkflow)
}
