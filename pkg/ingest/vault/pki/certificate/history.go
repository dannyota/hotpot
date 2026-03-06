package certificate

import (
	"context"
	"fmt"
	"time"

	entpki "danny.vn/hotpot/pkg/storage/ent/vault/pki"
	"danny.vn/hotpot/pkg/storage/ent/vault/pki/bronzehistoryvaultpkicertificate"
)

// HistoryService handles history tracking for certificates.
type HistoryService struct {
	entClient *entpki.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entpki.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new certificate.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entpki.Tx, data *CertificateData, now time.Time) error {
	create := tx.BronzeHistoryVaultPKICertificate.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetVaultName(data.VaultName).
		SetMountPath(data.MountPath).
		SetSerialNumber(data.SerialNumber).
		SetCommonName(data.CommonName).
		SetIssuerCn(data.IssuerCN).
		SetSubjectCn(data.SubjectCN).
		SetSans(data.SANs).
		SetKeyType(data.KeyType).
		SetKeyBits(data.KeyBits).
		SetSigningAlgo(data.SigningAlgo).
		SetNotBefore(data.NotBefore).
		SetNotAfter(data.NotAfter).
		SetIsRevoked(data.IsRevoked).
		SetCertificatePem(data.CertPEM)

	if data.RevokedAt != nil {
		create.SetRevokedAt(*data.RevokedAt)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create certificate history: %w", err)
	}

	return nil
}

// UpdateHistory closes old history and creates new history for a changed certificate.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entpki.Tx, old *entpki.BronzeVaultPKICertificate, new *CertificateData, now time.Time) error {
	// Find current open history record
	currentHist, err := tx.BronzeHistoryVaultPKICertificate.Query().
		Where(
			bronzehistoryvaultpkicertificate.ResourceID(old.ID),
			bronzehistoryvaultpkicertificate.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current certificate history: %w", err)
	}

	// Close old history
	if err := tx.BronzeHistoryVaultPKICertificate.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close certificate history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryVaultPKICertificate.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetVaultName(new.VaultName).
		SetMountPath(new.MountPath).
		SetSerialNumber(new.SerialNumber).
		SetCommonName(new.CommonName).
		SetIssuerCn(new.IssuerCN).
		SetSubjectCn(new.SubjectCN).
		SetSans(new.SANs).
		SetKeyType(new.KeyType).
		SetKeyBits(new.KeyBits).
		SetSigningAlgo(new.SigningAlgo).
		SetNotBefore(new.NotBefore).
		SetNotAfter(new.NotAfter).
		SetIsRevoked(new.IsRevoked).
		SetCertificatePem(new.CertPEM)

	if new.RevokedAt != nil {
		create.SetRevokedAt(*new.RevokedAt)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new certificate history: %w", err)
	}

	return nil
}

// CloseHistory closes the history record for a deleted certificate.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entpki.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryVaultPKICertificate.Query().
		Where(
			bronzehistoryvaultpkicertificate.ResourceID(resourceID),
			bronzehistoryvaultpkicertificate.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entpki.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current certificate history: %w", err)
	}

	if err := tx.BronzeHistoryVaultPKICertificate.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close certificate history: %w", err)
	}

	return nil
}
