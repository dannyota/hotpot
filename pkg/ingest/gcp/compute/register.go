package compute

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/compute/address"
	"hotpot/pkg/ingest/gcp/compute/disk"
	"hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"hotpot/pkg/ingest/gcp/compute/globaladdress"
	"hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"hotpot/pkg/ingest/gcp/compute/healthcheck"
	"hotpot/pkg/ingest/gcp/compute/image"
	"hotpot/pkg/ingest/gcp/compute/instance"
	"hotpot/pkg/ingest/gcp/compute/instancegroup"
	"hotpot/pkg/ingest/gcp/compute/network"
	"hotpot/pkg/ingest/gcp/compute/snapshot"
	"hotpot/pkg/ingest/gcp/compute/subnetwork"
	"hotpot/pkg/ingest/gcp/compute/targetinstance"
	"hotpot/pkg/ingest/gcp/compute/targetvpngateway"
	"hotpot/pkg/ingest/gcp/compute/vpngateway"
	"hotpot/pkg/ingest/gcp/compute/vpntunnel"
	"hotpot/pkg/storage/ent"
)

// Register registers all Compute Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	instance.Register(w, configService, entClient, limiter)
	disk.Register(w, configService, entClient, limiter)
	network.Register(w, configService, entClient, limiter)
	subnetwork.Register(w, configService, entClient, limiter)
	instancegroup.Register(w, configService, entClient, limiter)
	snapshot.Register(w, configService, entClient, limiter)
	targetinstance.Register(w, configService, entClient, limiter)
	address.Register(w, configService, entClient, limiter)
	globaladdress.Register(w, configService, entClient, limiter)
	forwardingrule.Register(w, configService, entClient, limiter)
	globalforwardingrule.Register(w, configService, entClient, limiter)
	healthcheck.Register(w, configService, entClient, limiter)
	image.Register(w, configService, entClient, limiter)
	vpngateway.Register(w, configService, entClient, limiter)
	targetvpngateway.Register(w, configService, entClient, limiter)
	vpntunnel.Register(w, configService, entClient, limiter)

	// Register compute workflow
	w.RegisterWorkflow(GCPComputeWorkflow)
}
