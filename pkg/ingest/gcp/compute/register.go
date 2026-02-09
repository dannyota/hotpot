package compute

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/address"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/backendservice"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/disk"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globaladdress"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/healthcheck"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/image"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instance"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/instancegroup"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/neg"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/negendpoint"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/network"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/snapshot"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/subnetwork"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targethttpproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targethttpsproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetinstance"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetpool"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetsslproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targettcpproxy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/targetvpngateway"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/urlmap"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/vpngateway"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute/vpntunnel"
	"github.com/dannyota/hotpot/pkg/storage/ent"
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
	targethttpproxy.Register(w, configService, entClient, limiter)
	targettcpproxy.Register(w, configService, entClient, limiter)
	targetsslproxy.Register(w, configService, entClient, limiter)
	targethttpsproxy.Register(w, configService, entClient, limiter)
	urlmap.Register(w, configService, entClient, limiter)
	targetpool.Register(w, configService, entClient, limiter)
	neg.Register(w, configService, entClient, limiter)
	negendpoint.Register(w, configService, entClient, limiter)
	backendservice.Register(w, configService, entClient, limiter)

	// Register compute workflow
	w.RegisterWorkflow(GCPComputeWorkflow)
}
