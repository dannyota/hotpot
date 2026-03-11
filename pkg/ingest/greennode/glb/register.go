package glb

import (
	"danny.vn/gnode/auth"
	"entgo.io/ent/dialect"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/glb/glbpackage"
	"danny.vn/hotpot/pkg/ingest/greennode/glb/glbregion"
	"danny.vn/hotpot/pkg/ingest/greennode/glb/glbresource"
	entglb "danny.vn/hotpot/pkg/storage/ent/greennode/glb"
)

// Register registers all GreenNode GLB activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entglb.NewClient(entglb.Driver(driver), entglb.AlternateSchema(entglb.DefaultSchemaConfig()))
	glbresource.Register(w, configService, entClient, iamAuth, limiter)
	glbpackage.Register(w, configService, entClient, iamAuth, limiter)
	glbregion.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeGLBWorkflow)
}
