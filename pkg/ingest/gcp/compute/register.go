package compute

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

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
)

// Register registers all Compute Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	// Register sub-packages with config service
	instance.Register(w, configService, db, limiter)
	disk.Register(w, configService, db, limiter)
	network.Register(w, configService, db, limiter)
	subnetwork.Register(w, configService, db, limiter)
	instancegroup.Register(w, configService, db, limiter)
	snapshot.Register(w, configService, db, limiter)
	targetinstance.Register(w, configService, db, limiter)
	address.Register(w, configService, db, limiter)
	globaladdress.Register(w, configService, db, limiter)
	forwardingrule.Register(w, configService, db, limiter)
	globalforwardingrule.Register(w, configService, db, limiter)
	healthcheck.Register(w, configService, db, limiter)
	image.Register(w, configService, db, limiter)
	vpngateway.Register(w, configService, db, limiter)
	targetvpngateway.Register(w, configService, db, limiter)
	vpntunnel.Register(w, configService, db, limiter)

	// Register compute workflow
	w.RegisterWorkflow(GCPComputeWorkflow)
}
