package certificate

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeloadbalancercertificate"
)

// Service handles GreenNode certificate ingestion.
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

// IngestResult contains the result of certificate ingestion.
type IngestResult struct {
	CertificateCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

// Ingest fetches certificates from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	certs, err := s.client.ListCertificates(ctx)
	if err != nil {
		return nil, fmt.Errorf("list certificates: %w", err)
	}

	dataList := make([]*CertificateData, 0, len(certs))
	for i := range certs {
		dataList = append(dataList, ConvertCertificate(&certs[i], projectID, region, collectedAt))
	}

	if err := s.saveCertificates(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save certificates: %w", err)
	}

	return &IngestResult{
		CertificateCount: len(dataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

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

	for _, data := range certs {
		existing, err := tx.BronzeGreenNodeLoadBalancerCertificate.Query().
			Where(bronzegreennodeloadbalancercertificate.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing certificate %s: %w", data.Name, err)
		}

		diff := DiffCertificateData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeLoadBalancerCertificate.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for certificate %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			_, err = tx.BronzeGreenNodeLoadBalancerCertificate.Create().
				SetID(data.ID).
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
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create certificate %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for certificate %s: %w", data.Name, err)
			}
		} else {
			_, err = tx.BronzeGreenNodeLoadBalancerCertificate.UpdateOneID(data.ID).
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
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update certificate %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for certificate %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleCertificates removes certificates not collected in the latest run.
func (s *Service) DeleteStaleCertificates(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeLoadBalancerCertificate.Query().
		Where(
			bronzegreennodeloadbalancercertificate.ProjectID(projectID),
			bronzegreennodeloadbalancercertificate.Region(region),
			bronzegreennodeloadbalancercertificate.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale certificates: %w", err)
	}

	for _, c := range stale {
		if err := s.history.CloseHistory(ctx, tx, c.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for certificate %s: %w", c.ID, err)
		}
		if err := tx.BronzeGreenNodeLoadBalancerCertificate.DeleteOneID(c.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete certificate %s: %w", c.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
