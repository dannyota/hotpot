package targetsslproxy

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputetargetsslproxy"
)

// Service handles GCP Compute target SSL proxy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target SSL proxy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target SSL proxy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target SSL proxy ingestion.
type IngestResult struct {
	ProjectID           string
	TargetSslProxyCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches target SSL proxies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	proxies, err := s.client.ListTargetSslProxies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target SSL proxies: %w", err)
	}

	proxyDataList := make([]*TargetSslProxyData, 0, len(proxies))
	for _, p := range proxies {
		data, err := ConvertTargetSslProxy(p, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert target SSL proxy: %w", err)
		}
		proxyDataList = append(proxyDataList, data)
	}

	if err := s.saveTargetSslProxies(ctx, proxyDataList); err != nil {
		return nil, fmt.Errorf("failed to save target SSL proxies: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		TargetSslProxyCount: len(proxyDataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveTargetSslProxies(ctx context.Context, proxies []*TargetSslProxyData) error {
	if len(proxies) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, proxyData := range proxies {
		existing, err := tx.BronzeGCPComputeTargetSslProxy.Query().
			Where(bronzegcpcomputetargetsslproxy.ID(proxyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing target SSL proxy %s: %w", proxyData.Name, err)
		}

		diff := DiffTargetSslProxyData(existing, proxyData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeTargetSslProxy.UpdateOneID(proxyData.ID).
				SetCollectedAt(proxyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for target SSL proxy %s: %w", proxyData.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeTargetSslProxy.Create().
				SetID(proxyData.ID).
				SetName(proxyData.Name).
				SetProjectID(proxyData.ProjectID).
				SetCollectedAt(proxyData.CollectedAt).
				SetFirstCollectedAt(proxyData.CollectedAt)

			if proxyData.Description != "" {
				create.SetDescription(proxyData.Description)
			}
			if proxyData.CreationTimestamp != "" {
				create.SetCreationTimestamp(proxyData.CreationTimestamp)
			}
			if proxyData.SelfLink != "" {
				create.SetSelfLink(proxyData.SelfLink)
			}
			if proxyData.Service != "" {
				create.SetService(proxyData.Service)
			}
			if proxyData.ProxyHeader != "" {
				create.SetProxyHeader(proxyData.ProxyHeader)
			}
			if proxyData.CertificateMap != "" {
				create.SetCertificateMap(proxyData.CertificateMap)
			}
			if proxyData.SslPolicy != "" {
				create.SetSslPolicy(proxyData.SslPolicy)
			}
			if proxyData.SslCertificatesJSON != nil {
				create.SetSslCertificatesJSON(proxyData.SslCertificatesJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create target SSL proxy %s: %w", proxyData.Name, err)
			}
		} else {
			update := tx.BronzeGCPComputeTargetSslProxy.UpdateOneID(proxyData.ID).
				SetName(proxyData.Name).
				SetProjectID(proxyData.ProjectID).
				SetCollectedAt(proxyData.CollectedAt)

			if proxyData.Description != "" {
				update.SetDescription(proxyData.Description)
			}
			if proxyData.CreationTimestamp != "" {
				update.SetCreationTimestamp(proxyData.CreationTimestamp)
			}
			if proxyData.SelfLink != "" {
				update.SetSelfLink(proxyData.SelfLink)
			}
			if proxyData.Service != "" {
				update.SetService(proxyData.Service)
			}
			if proxyData.ProxyHeader != "" {
				update.SetProxyHeader(proxyData.ProxyHeader)
			}
			if proxyData.CertificateMap != "" {
				update.SetCertificateMap(proxyData.CertificateMap)
			}
			if proxyData.SslPolicy != "" {
				update.SetSslPolicy(proxyData.SslPolicy)
			}
			if proxyData.SslCertificatesJSON != nil {
				update.SetSslCertificatesJSON(proxyData.SslCertificatesJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update target SSL proxy %s: %w", proxyData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, proxyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for target SSL proxy %s: %w", proxyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, proxyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for target SSL proxy %s: %w", proxyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetSslProxies removes target SSL proxies not collected in the latest run.
func (s *Service) DeleteStaleTargetSslProxies(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleProxies, err := tx.BronzeGCPComputeTargetSslProxy.Query().
		Where(
			bronzegcpcomputetargetsslproxy.ProjectID(projectID),
			bronzegcpcomputetargetsslproxy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, proxy := range staleProxies {
		if err := s.history.CloseHistory(ctx, tx, proxy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for target SSL proxy %s: %w", proxy.ID, err)
		}

		if err := tx.BronzeGCPComputeTargetSslProxy.DeleteOne(proxy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete target SSL proxy %s: %w", proxy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
