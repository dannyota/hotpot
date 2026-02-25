package pubsub

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/subscription"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub/topic"
	"entgo.io/ent/dialect"
	entpubsub "github.com/dannyota/hotpot/pkg/storage/ent/gcp/pubsub"
)

// Register registers all Pub/Sub activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entpubsub.NewClient(entpubsub.Driver(driver), entpubsub.AlternateSchema(entpubsub.DefaultSchemaConfig()))
	topic.Register(w, configService, entClient, limiter)
	subscription.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPPubSubWorkflow)
}
