package glb

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbpackage"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbregion"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbresource"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode GLB activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	glbresource.Register(w, configService, entClient, iamAuth, limiter)
	glbpackage.Register(w, configService, entClient, iamAuth, limiter)
	glbregion.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeGLBWorkflow)
}
