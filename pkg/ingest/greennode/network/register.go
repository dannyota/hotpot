package network

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/endpoint"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/secgroup"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode network activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	secgroup.Register(w, configService, entClient, iamAuth, limiter)
	endpoint.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeNetworkWorkflow)
}
