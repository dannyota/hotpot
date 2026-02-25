package compute

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/osimage"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/server"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/servergroup"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/sshkey"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/compute/userimage"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode compute activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	server.Register(w, configService, entClient, iamAuth, limiter)
	sshkey.Register(w, configService, entClient, iamAuth, limiter)
	servergroup.Register(w, configService, entClient, iamAuth, limiter)
	osimage.Register(w, configService, entClient, iamAuth, limiter)
	userimage.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeComputeWorkflow)
}
