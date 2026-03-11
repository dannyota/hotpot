package subscription

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	entpubsub "danny.vn/hotpot/pkg/storage/ent/gcp/pubsub"
)

// Register registers all Pub/Sub subscription activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *entpubsub.Client, limiter ratelimit.Limiter) {
	activities := NewActivities(configService, entClient, limiter)
	w.RegisterActivity(activities.IngestPubSubSubscriptions)
	w.RegisterWorkflow(GCPPubSubSubscriptionWorkflow)
}
