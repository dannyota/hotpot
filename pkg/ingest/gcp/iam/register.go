package iam

import (
	"go.temporal.io/sdk/worker"

	"hotpot/pkg/base/config"
	"hotpot/pkg/base/ratelimit"
	"hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
	"hotpot/pkg/storage/ent"
)

// Register registers all IAM activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	serviceaccount.Register(w, configService, entClient, limiter)
	serviceaccountkey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAMWorkflow)
}
