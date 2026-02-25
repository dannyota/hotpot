package certificate

import (
	"context"
	"fmt"
	"time"

	entlb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/loadbalancer/bronzehistorygreennodeloadbalancercertificate"
)

// HistoryService handles history tracking for certificates.
type HistoryService struct {
	entClient *entlb.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entlb.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new certificate.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entlb.Tx, data *CertificateData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeLoadBalancerCertificate.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCertificateType(data.CertificateType).
		SetExpiredAt(data.ExpiredAt).
		SetImportedAt(data.ImportedAt).
		SetNotAfter(data.NotAfter).
		SetKeyAlgorithm(data.KeyAlgorithm).
		SetSerial(data.Serial).
		SetSubject(data.Subject).
		SetDomainName(data.DomainName).
		SetInUse(data.InUse).
		SetIssuer(data.Issuer).
		SetSignatureAlgorithm(data.SignatureAlgorithm).
		SetNotBefore(data.NotBefore).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create certificate history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entlb.Tx, old *entlb.BronzeGreenNodeLoadBalancerCertificate, new *CertificateData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerCertificate.Query().
		Where(
			bronzehistorygreennodeloadbalancercertificate.ResourceID(old.ID),
			bronzehistorygreennodeloadbalancercertificate.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current certificate history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeLoadBalancerCertificate.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close certificate history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeLoadBalancerCertificate.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetCertificateType(new.CertificateType).
		SetExpiredAt(new.ExpiredAt).
		SetImportedAt(new.ImportedAt).
		SetNotAfter(new.NotAfter).
		SetKeyAlgorithm(new.KeyAlgorithm).
		SetSerial(new.Serial).
		SetSubject(new.Subject).
		SetDomainName(new.DomainName).
		SetInUse(new.InUse).
		SetIssuer(new.Issuer).
		SetSignatureAlgorithm(new.SignatureAlgorithm).
		SetNotBefore(new.NotBefore).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new certificate history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted certificate.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entlb.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerCertificate.Query().
		Where(
			bronzehistorygreennodeloadbalancercertificate.ResourceID(resourceID),
			bronzehistorygreennodeloadbalancercertificate.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entlb.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current certificate history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeLoadBalancerCertificate.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close certificate history: %w", err)
	}
	return nil
}
