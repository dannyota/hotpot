package certificate

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzevaultpkicertificate"
)

// Service handles Vault PKI certificate ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new certificate ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for certificate ingestion.
type IngestParams struct {
	VaultName string
	MountPath string
}

// IngestResult contains the result of certificate ingestion.
type IngestResult struct {
	CertificateCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches certificates from a Vault PKI mount and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// List all certificate serials
	serials, err := s.client.ListCertSerials(ctx, params.MountPath)
	if err != nil {
		return nil, fmt.Errorf("list cert serials: %w", err)
	}

	// Fetch and convert each certificate
	certs := make([]*CertificateData, 0, len(serials))
	for _, serial := range serials {
		resp, err := s.client.GetCert(ctx, params.MountPath, serial)
		if err != nil {
			return nil, fmt.Errorf("get cert %s: %w", serial, err)
		}

		data, err := ConvertCert(resp, params.VaultName, params.MountPath, serial, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert cert %s: %w", serial, err)
		}
		certs = append(certs, data)
	}

	// Save to database
	if err := s.saveCertificates(ctx, certs); err != nil {
		return nil, fmt.Errorf("save certificates: %w", err)
	}

	return &IngestResult{
		CertificateCount: len(certs),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveCertificates saves certificates to the database with history tracking.
func (s *Service) saveCertificates(ctx context.Context, certs []*CertificateData) error {
	if len(certs) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, certData := range certs {
		// Load existing certificate
		existing, err := tx.BronzeVaultPKICertificate.Query().
			Where(bronzevaultpkicertificate.ID(certData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing cert %s: %w", certData.ID, err)
		}

		// Compute diff
		diff := DiffCertData(existing, certData)

		// Skip if no changes
		if !diff.IsNew && !diff.IsChanged && existing != nil {
			// Update collected_at only
			if err := tx.BronzeVaultPKICertificate.UpdateOneID(certData.ID).
				SetCollectedAt(certData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for cert %s: %w", certData.ID, err)
			}
			continue
		}

		if existing == nil {
			// Create new certificate
			create := tx.BronzeVaultPKICertificate.Create().
				SetID(certData.ID).
				SetVaultName(certData.VaultName).
				SetMountPath(certData.MountPath).
				SetSerialNumber(certData.SerialNumber).
				SetCommonName(certData.CommonName).
				SetIssuerCn(certData.IssuerCN).
				SetSubjectCn(certData.SubjectCN).
				SetSans(certData.SANs).
				SetKeyType(certData.KeyType).
				SetKeyBits(certData.KeyBits).
				SetSigningAlgo(certData.SigningAlgo).
				SetNotBefore(certData.NotBefore).
				SetNotAfter(certData.NotAfter).
				SetIsRevoked(certData.IsRevoked).
				SetCertificatePem(certData.CertPEM).
				SetCollectedAt(certData.CollectedAt).
				SetFirstCollectedAt(certData.CollectedAt)

			if certData.RevokedAt != nil {
				create.SetRevokedAt(*certData.RevokedAt)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create cert %s: %w", certData.ID, err)
			}
		} else {
			// Update existing certificate
			update := tx.BronzeVaultPKICertificate.UpdateOneID(certData.ID).
				SetVaultName(certData.VaultName).
				SetMountPath(certData.MountPath).
				SetSerialNumber(certData.SerialNumber).
				SetCommonName(certData.CommonName).
				SetIssuerCn(certData.IssuerCN).
				SetSubjectCn(certData.SubjectCN).
				SetSans(certData.SANs).
				SetKeyType(certData.KeyType).
				SetKeyBits(certData.KeyBits).
				SetSigningAlgo(certData.SigningAlgo).
				SetNotBefore(certData.NotBefore).
				SetNotAfter(certData.NotAfter).
				SetIsRevoked(certData.IsRevoked).
				SetCertificatePem(certData.CertPEM).
				SetCollectedAt(certData.CollectedAt)

			if certData.RevokedAt != nil {
				update.SetRevokedAt(*certData.RevokedAt)
			} else {
				update.ClearRevokedAt()
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update cert %s: %w", certData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, certData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for cert %s: %w", certData.ID, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, certData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for cert %s: %w", certData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes certificates that were not collected in the latest run.
// Scoped to a specific vault_name + mount_path pair.
func (s *Service) DeleteStale(ctx context.Context, vaultName, mountPath string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	// Find stale certificates
	staleCerts, err := tx.BronzeVaultPKICertificate.Query().
		Where(
			bronzevaultpkicertificate.VaultName(vaultName),
			bronzevaultpkicertificate.MountPath(mountPath),
			bronzevaultpkicertificate.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale certificate
	for _, cert := range staleCerts {
		if err := s.history.CloseHistory(ctx, tx, cert.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for cert %s: %w", cert.ID, err)
		}

		if err := tx.BronzeVaultPKICertificate.DeleteOne(cert).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete cert %s: %w", cert.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
