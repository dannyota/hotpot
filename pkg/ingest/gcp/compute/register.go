package compute

import (
	"go.temporal.io/sdk/worker"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/ratelimit"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/address"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/backendservice"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/disk"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/firewall"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/forwardingrule"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/globaladdress"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/globalforwardingrule"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/healthcheck"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/image"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/instance"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/interconnect"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/instancegroup"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/neg"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/negendpoint"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/network"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/packetmirroring"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/projectmetadata"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/router"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/securitypolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/snapshot"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/sslpolicy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/subnetwork"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targethttpproxy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targethttpsproxy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targetinstance"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targetpool"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targetsslproxy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targettcpproxy"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/targetvpngateway"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/urlmap"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/vpngateway"
	"danny.vn/hotpot/pkg/ingest/gcp/compute/vpntunnel"
	"entgo.io/ent/dialect"
	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
	entvpn "danny.vn/hotpot/pkg/storage/ent/gcp/vpn"
)

// Register registers all Compute Engine activities and workflows.
// Client is NOT created here - it's created per workflow session.
func Register(w worker.Worker, configService *config.Service, driver dialect.Driver, limiter ratelimit.Limiter) {
	entClient := entcompute.NewClient(entcompute.Driver(driver), entcompute.AlternateSchema(entcompute.DefaultSchemaConfig()))
	vpnClient := entvpn.NewClient(entvpn.Driver(driver), entvpn.AlternateSchema(entvpn.DefaultSchemaConfig()))
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
	vpngateway.Register(w, configService, vpnClient, limiter)
	targetvpngateway.Register(w, configService, vpnClient, limiter)
	vpntunnel.Register(w, configService, vpnClient, limiter)
	targethttpproxy.Register(w, configService, entClient, limiter)
	targettcpproxy.Register(w, configService, entClient, limiter)
	targetsslproxy.Register(w, configService, entClient, limiter)
	targethttpsproxy.Register(w, configService, entClient, limiter)
	urlmap.Register(w, configService, entClient, limiter)
	targetpool.Register(w, configService, entClient, limiter)
	neg.Register(w, configService, entClient, limiter)
	negendpoint.Register(w, configService, entClient, limiter)
	backendservice.Register(w, configService, entClient, limiter)
	firewall.Register(w, configService, entClient, limiter)
	sslpolicy.Register(w, configService, entClient, limiter)
	router.Register(w, configService, entClient, limiter)
	securitypolicy.Register(w, configService, entClient, limiter)
	interconnect.Register(w, configService, entClient, limiter)
	packetmirroring.Register(w, configService, entClient, limiter)
	projectmetadata.Register(w, configService, entClient, limiter)

	// Register compute workflow
	w.RegisterWorkflow(GCPComputeWorkflow)
}
