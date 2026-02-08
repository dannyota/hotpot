package gcp

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/compute"
	"hotpot/pkg/ingest/gcp/container"
	"hotpot/pkg/ingest/gcp/iam"
	"hotpot/pkg/ingest/gcp/resourcemanager"
	"hotpot/pkg/ingest/gcp/vpcaccess"
	"hotpot/pkg/storage/ent"
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

	// Register GCP inventory workflow
	w.RegisterWorkflow(GCPInventoryWorkflow)

	return rateLimitSvc
}
