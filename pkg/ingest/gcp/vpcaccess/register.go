package vpcaccess

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/vpcaccess/connector"
)

// Register registers all VPC Access activities and workflows.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	connector.Register(w, configService, db, limiter)

	w.RegisterWorkflow(GCPVpcAccessWorkflow)
}
