package pubsub

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/pubsub/subscription"
	"danny.vn/hotpot/pkg/ingest/gcp/pubsub/topic"
	"entgo.io/ent/dialect"
	entpubsub "danny.vn/hotpot/pkg/storage/ent/gcp/pubsub"
)

// Register registers all Pub/Sub activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entpubsub.NewClient(entpubsub.Driver(driver), entpubsub.AlternateSchema(entpubsub.DefaultSchemaConfig()))
	topic.Register(w, configService, entClient, limiter)
	subscription.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPPubSubWorkflow)
}
