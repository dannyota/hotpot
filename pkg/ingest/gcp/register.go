package gcp

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/accesscontextmanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/alloydb"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/appengine"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigquery"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/bigtable"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/binaryauthorization"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudasset"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/cloudfunctions"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/compute"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/container"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/containeranalysis"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dataproc"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/dns"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/filestore"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iap"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/kms"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/logging"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/monitoring"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/orgpolicy"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/pubsub"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/redis"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/resourcemanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/run"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/secretmanager"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/securitycenter"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/serviceusage"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/spanner"
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

	// Register Security Command Center (sources, findings, notification configs)
	securitycenter.Register(w, configService, entClient, limiter)

	// Register Organization Policy (constraints, policies, custom constraints)
	orgpolicy.Register(w, configService, entClient, limiter)

	// Register Service Usage (enabled services)
	serviceusage.Register(w, configService, entClient, limiter)

	// Register Cloud Functions
	cloudfunctions.Register(w, configService, entClient, limiter)

	// Register Memorystore Redis (instances)
	redis.Register(w, configService, entClient, limiter)

	// Register Dataproc (clusters)
	dataproc.Register(w, configService, entClient, limiter)

	// Register IAP (settings, IAM policies)
	iap.Register(w, configService, entClient, limiter)

	// Register AlloyDB (clusters)
	alloydb.Register(w, configService, entClient, limiter)

	// Register Filestore (instances)
	filestore.Register(w, configService, entClient, limiter)

	// Register Pub/Sub (topics, subscriptions)
	pubsub.Register(w, configService, entClient, limiter)

	// Register App Engine (applications, services)
	appengine.Register(w, configService, entClient, limiter)

	// Register Cloud Asset (assets, IAM policy search, resource search)
	cloudasset.Register(w, configService, entClient, limiter)

	// Register Binary Authorization (policies, attestors)
	binaryauthorization.Register(w, configService, entClient, limiter)

	// Register Monitoring (alert policies, uptime checks)
	monitoring.Register(w, configService, entClient, limiter)

	// Register Cloud Run (services, revisions)
	run.Register(w, configService, entClient, limiter)

	// Register Access Context Manager (access policies, levels, perimeters)
	accesscontextmanager.Register(w, configService, entClient, limiter)

	// Register Container Analysis (notes, occurrences)
	containeranalysis.Register(w, configService, entClient, limiter)

	// Register Spanner (instances, databases)
	spanner.Register(w, configService, entClient, limiter)

	// Register BigQuery (datasets, tables)
	bigquery.Register(w, configService, entClient, limiter)

	// Register Bigtable (instances, clusters)
	bigtable.Register(w, configService, entClient, limiter)

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)

	return rateLimitSvc
}
