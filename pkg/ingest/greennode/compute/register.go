package compute

import (
	"entgo.io/ent/dialect"
	"danny.vn/gnode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/osimage"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/server"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/servergroup"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/sshkey"
	"danny.vn/hotpot/pkg/ingest/greennode/compute/userimage"
	entcompute "danny.vn/hotpot/pkg/storage/ent/greennode/compute"
)

// Register registers all GreenNode compute activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entcompute.NewClient(entcompute.Driver(driver), entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()))
	server.Register(w, configService, entClient, iamAuth, limiter)
	sshkey.Register(w, configService, entClient, iamAuth, limiter)
	servergroup.Register(w, configService, entClient, iamAuth, limiter)
	osimage.Register(w, configService, entClient, iamAuth, limiter)
	userimage.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeComputeWorkflow)
}
