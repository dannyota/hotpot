package compute

import (
	"entgo.io/ent/dialect"
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/osimage"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/server"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/servergroup"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/sshkey"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/userimage"
	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
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
