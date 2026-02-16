package appengine

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine/application"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine/appservice"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all App Engine activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	application.Register(w, configService, entClient, limiter)
	appservice.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPAppEngineWorkflow)
}
