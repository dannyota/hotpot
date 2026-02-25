package glb

import (
	"danny.vn/greennode/auth"
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbpackage"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbregion"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/glb/glbresource"
	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
)

// Register registers all GreenNode GLB activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entglb.NewClient(entglb.Driver(driver), entglb.AlternateSchema(entglb.DefaultSchemaConfig()))
	glbresource.Register(w, configService, entClient, iamAuth, limiter)
	glbpackage.Register(w, configService, entClient, iamAuth, limiter)
	glbregion.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeGLBWorkflow)
}
