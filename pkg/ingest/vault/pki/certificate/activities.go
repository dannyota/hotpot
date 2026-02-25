package certificate

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/dannyota/hotpot/pkg/base/config"
	"github.com/dannyota/hotpot/pkg/base/ratelimit"
	entpki "github.com/dannyota/hotpot/pkg/storage/ent/vault/pki"
)

// Activities holds dependencies for Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entpki.Client
	limiter       ratelimit.Limiter
}

// NewActivities creates a new Activities instance.
func NewActivities(configService *config.Service, entClient *entpki.Client, limiter ratelimit.Limiter) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
		limiter:       limiter,
	}
}

// IngestCertificatesParams contains parameters for the ingest activity.
type IngestCertificatesParams struct {
	VaultName string
	MountPath string
}

// IngestCertificatesResult contains the result of the ingest activity.
type IngestCertificatesResult struct {
	CertificateCount int
	DurationMillis   int64
}

// IngestCertificatesActivity is the activity function reference for workflow registration.
var IngestCertificatesActivity = (*Activities).IngestCertificates

// IngestCertificates is a Temporal activity that ingests PKI certificates from a Vault mount.
func (a *Activities) IngestCertificates(ctx context.Context, params IngestCertificatesParams) (*IngestCertificatesResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting Vault PKI certificate ingestion",
		"vaultName", params.VaultName,
		"mountPath", params.MountPath,
	)

	// Look up vault instance config
	inst := a.configService.VaultInstance(params.VaultName)
	if inst == nil {
		return nil, fmt.Errorf("vault instance %q not found in config", params.VaultName)
	}

	// Create client
	verifySSL := true
	if inst.VerifySSL != nil {
		verifySSL = *inst.VerifySSL
	}
	client := NewClient(inst.Address, inst.Token, verifySSL, a.limiter)

	// Create service and ingest
	service := NewService(client, a.entClient)
	result, err := service.Ingest(ctx, IngestParams{
		VaultName: params.VaultName,
		MountPath: params.MountPath,
	})
	if err != nil {
		return nil, fmt.Errorf("ingest certificates: %w", err)
	}

	// Delete stale certificates
	if err := service.DeleteStale(ctx, params.VaultName, params.MountPath, result.CollectedAt); err != nil {
		logger.Warn("Failed to delete stale certificates", "error", err)
	}

	logger.Info("Completed Vault PKI certificate ingestion",
		"vaultName", params.VaultName,
		"mountPath", params.MountPath,
		"certificateCount", result.CertificateCount,
		"durationMillis", result.DurationMillis,
	)

	return &IngestCertificatesResult{
		CertificateCount: result.CertificateCount,
		DurationMillis:   result.DurationMillis,
	}, nil
}
