package iam

import (
	"go.temporal.io/sdk/worker"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccount"
	"github.com/dannyota/hotpot/pkg/ingest/gcp/iam/serviceaccountkey"
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// Register registers all IAM activities and workflows.
func Register(w worker.Worker, configService *config.Service, entClient *ent.Client, limiter ratelimit.Limiter) {
	serviceaccount.Register(w, configService, entClient, limiter)
	serviceaccountkey.Register(w, configService, entClient, limiter)

	w.RegisterWorkflow(GCPIAMWorkflow)
}
