package network

import (
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/endpoint"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/interconnect"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/peering"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/routetable"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/secgroup"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/subnet"
	"github.com/dannyota/hotpot/pkg/ingest/greennode/network/vpc"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GreenNode network activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	secgroup.Register(w, configService, entClient, iamAuth, limiter)
	endpoint.Register(w, configService, entClient, iamAuth, limiter)
	vpc.Register(w, configService, entClient, iamAuth, limiter)
	subnet.Register(w, configService, entClient, iamAuth, limiter)
	routetable.Register(w, configService, entClient, iamAuth, limiter)
	peering.Register(w, configService, entClient, iamAuth, limiter)
	interconnect.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeNetworkWorkflow)
}
