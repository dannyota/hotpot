package iam

import (
	"go.temporal.io/sdk/worker"
	"gorm.io/gorm"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
)

// Register registers all IAM activities and workflows.
func Register(w worker.Worker, configService *config.Service, db *gorm.DB, limiter ratelimit.Limiter) {
	serviceaccount.Register(w, configService, db, limiter)
	serviceaccountkey.Register(w, configService, db, limiter)

	w.RegisterWorkflow(GCPIAMWorkflow)
}
