package accesscontextmanager

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesslevel"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/accesspolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager/serviceperimeter"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all Access Context Manager activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	accesspolicy.Register(w, configService, entClient, limiter)
	accesslevel.Register(w, configService, entClient, limiter)
	serviceperimeter.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPAccessContextManagerWorkflow)
}
