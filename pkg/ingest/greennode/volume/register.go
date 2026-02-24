package volume

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/blockvolume"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/volumetype"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/volume/volumetypezone"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode volume activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	blockvolume.Register(w, configService, entClient, iamAuth, limiter)
	volumetype.Register(w, configService, entClient, iamAuth, limiter)
	volumetypezone.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeVolumeWorkflow)
}
