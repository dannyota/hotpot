package gcp

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/container"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager"
	gcpsql "github.com/dannyota/hotpot/pkg/ingest/gcp/sql"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/storage"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/vpcaccess"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all GCP activities and workflows with the Temporal worker.
// Returns the rate limit service for cleanup (caller should defer Close()).
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client) *ratelimit.Service {
	// Create shared rate limiter for all GCP API calls
	rateLimitSvc := ratelimit.NewService(ratelimit.ServiceOptions{
		RedisConfig: configService.RedisConfig(),
		KeyPrefix:   "ratelimit:gcp",
		ReqPerMin:   configService.GCPRateLimitPerMinute(),
	})
	limiter := rateLimitSvc.Limiter()

	// Register resource manager (project discovery)
	resourcemanager.Register(w, configService, entClient, limiter)

	// Register compute (instances, disks, networks)
	compute.Register(w, configService, entClient, limiter)

	// Register container (GKE clusters)
	container.Register(w, configService, entClient, limiter)

	// Register IAM (service accounts, keys)
	iam.Register(w, configService, entClient, limiter)

	// Register VPC Access (connectors)
	vpcaccess.Register(w, configService, entClient, limiter)

	// Register Storage (buckets)
	storage.Register(w, configService, entClient, limiter)

	// Register KMS (key rings, crypto keys)
	kms.Register(w, configService, entClient, limiter)

	// Register Logging (sinks, buckets)
	logging.Register(w, configService, entClient, limiter)

	// Register DNS (managed zones)
	dns.Register(w, configService, entClient, limiter)

	// Register Secret Manager (secrets)
	secretmanager.Register(w, configService, entClient, limiter)

	// Register Cloud SQL (instances)
	gcpsql.Register(w, configService, entClient, limiter)

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)

	return rateLimitSvc
}
