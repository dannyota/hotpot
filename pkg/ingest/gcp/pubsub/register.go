package pubsub

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/subscription"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/topic"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Pub/Sub activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	topic.Register(w, configService, entClient, limiter)
	subscription.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPPubSubWorkflow)
}
