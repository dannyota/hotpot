package network

import (
	"entgo.io/ent/dialect"
	"danny.vn/greennode/auth"
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/greennode/network/endpoint"
	"danny.vn/hotpot/pkg/ingest/greennode/network/interconnect"
	"danny.vn/hotpot/pkg/ingest/greennode/network/peering"
	"danny.vn/hotpot/pkg/ingest/greennode/network/routetable"
	"danny.vn/hotpot/pkg/ingest/greennode/network/secgroup"
	"danny.vn/hotpot/pkg/ingest/greennode/network/subnet"
	"danny.vn/hotpot/pkg/ingest/greennode/network/vpc"
	entnet "danny.vn/hotpot/pkg/storage/ent/greennode/network"
)

// Register registers all GreenNode network activities and workflows.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, iamAuth *auth.IAMUserAuth, limiter ratelimit.Limiter) {
	entClient := entnet.NewClient(entnet.Driver(driver), entnet.AlternateSchema(entnet.DefaultSchemaConfig()))
	secgroup.Register(w, configService, entClient, iamAuth, limiter)
	endpoint.Register(w, configService, entClient, iamAuth, limiter)
	vpc.Register(w, configService, entClient, iamAuth, limiter)
	subnet.Register(w, configService, entClient, iamAuth, limiter)
	routetable.Register(w, configService, entClient, iamAuth, limiter)
	peering.Register(w, configService, entClient, iamAuth, limiter)
	interconnect.Register(w, configService, entClient, iamAuth, limiter)

	w.RegisterWorkflow(GreenNodeNetworkWorkflow)
}
