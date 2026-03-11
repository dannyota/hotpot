package volume

import (
	"entgo.io/ent/dialect"
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/volume/blockvolume"
	"danny.vn/hotpot/pkg/ingest/greennode/volume/volumetype"
	"danny.vn/hotpot/pkg/ingest/greennode/volume/volumetypezone"
	entvol "danny.vn/hotpot/pkg/storage/ent/greennode/volume"
)

// Register registers all GreenNode volume activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entvol.NewClient(entvol.Driver(driver), entvol.AlternateSchema(entvol.DefaultSchemaConfig()))
	blockvolume.Register(w, configService, entClient, iamAuth, limiter)
	volumetype.Register(w, configService, entClient, iamAuth, limiter)
	volumetypezone.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeVolumeWorkflow)
}
