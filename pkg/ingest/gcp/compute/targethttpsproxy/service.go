package targethttpsproxy

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputetargethttpsproxy"
)

// Service handles GCP Compute target HTTPS proxy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target HTTPS proxy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target HTTPS proxy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target HTTPS proxy ingestion.
type IngestResult struct {
	ProjectID             string
	TargetHttpsProxyCount int
	CollectedAt           time.Time
	DurationMillis        int64
}

// Ingest fetches target HTTPS proxies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	proxies, err := s.client.ListTargetHttpsProxies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target HTTPS proxies: %w", err)
	}

	proxyDataList := make([]*TargetHttpsProxyData, 0, len(proxies))
	for _, p := range proxies {
		data, err := ConvertTargetHttpsProxy(p, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert target HTTPS proxy: %w", err)
		}
		proxyDataList = append(proxyDataList, data)
	}

	if err := s.saveTargetHttpsProxies(ctx, proxyDataList); err != nil {
		return nil, fmt.Errorf("failed to save target HTTPS proxies: %w", err)
	}

	return &IngestResult{
		ProjectID:             params.ProjectID,
		TargetHttpsProxyCount: len(proxyDataList),
		CollectedAt:           collectedAt,
		DurationMillis:        time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveTargetHttpsProxies(ctx context.Context, proxies []*TargetHttpsProxyData) error {
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
		existing, err := tx.BronzeGCPComputeTargetHttpsProxy.Query().
			Where(bronzegcpcomputetargethttpsproxy.ID(proxyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing target HTTPS proxy %s: %w", proxyData.Name, err)
		}

		diff := DiffTargetHttpsProxyData(existing, proxyData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeTargetHttpsProxy.UpdateOneID(proxyData.ID).
				SetCollectedAt(proxyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for target HTTPS proxy %s: %w", proxyData.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeTargetHttpsProxy.Create().
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
			if proxyData.Fingerprint != "" {
				create.SetFingerprint(proxyData.Fingerprint)
			}
			if proxyData.UrlMap != "" {
				create.SetURLMap(proxyData.UrlMap)
			}
			if proxyData.QuicOverride != "" {
				create.SetQuicOverride(proxyData.QuicOverride)
			}
			if proxyData.ServerTlsPolicy != "" {
				create.SetServerTLSPolicy(proxyData.ServerTlsPolicy)
			}
			if proxyData.AuthorizationPolicy != "" {
				create.SetAuthorizationPolicy(proxyData.AuthorizationPolicy)
			}
			if proxyData.CertificateMap != "" {
				create.SetCertificateMap(proxyData.CertificateMap)
			}
			if proxyData.SslPolicy != "" {
				create.SetSslPolicy(proxyData.SslPolicy)
			}
			if proxyData.TlsEarlyData != "" {
				create.SetTLSEarlyData(proxyData.TlsEarlyData)
			}
			if proxyData.ProxyBind {
				create.SetProxyBind(proxyData.ProxyBind)
			}
			if proxyData.HttpKeepAliveTimeoutSec != 0 {
				create.SetHTTPKeepAliveTimeoutSec(proxyData.HttpKeepAliveTimeoutSec)
			}
			if proxyData.SslCertificatesJSON != nil {
				create.SetSslCertificatesJSON(proxyData.SslCertificatesJSON)
			}
			if proxyData.Region != "" {
				create.SetRegion(proxyData.Region)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create target HTTPS proxy %s: %w", proxyData.Name, err)
			}
		} else {
			update := tx.BronzeGCPComputeTargetHttpsProxy.UpdateOneID(proxyData.ID).
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
			if proxyData.Fingerprint != "" {
				update.SetFingerprint(proxyData.Fingerprint)
			}
			if proxyData.UrlMap != "" {
				update.SetURLMap(proxyData.UrlMap)
			}
			if proxyData.QuicOverride != "" {
				update.SetQuicOverride(proxyData.QuicOverride)
			}
			if proxyData.ServerTlsPolicy != "" {
				update.SetServerTLSPolicy(proxyData.ServerTlsPolicy)
			}
			if proxyData.AuthorizationPolicy != "" {
				update.SetAuthorizationPolicy(proxyData.AuthorizationPolicy)
			}
			if proxyData.CertificateMap != "" {
				update.SetCertificateMap(proxyData.CertificateMap)
			}
			if proxyData.SslPolicy != "" {
				update.SetSslPolicy(proxyData.SslPolicy)
			}
			if proxyData.TlsEarlyData != "" {
				update.SetTLSEarlyData(proxyData.TlsEarlyData)
			}
			if proxyData.ProxyBind {
				update.SetProxyBind(proxyData.ProxyBind)
			}
			if proxyData.HttpKeepAliveTimeoutSec != 0 {
				update.SetHTTPKeepAliveTimeoutSec(proxyData.HttpKeepAliveTimeoutSec)
			}
			if proxyData.SslCertificatesJSON != nil {
				update.SetSslCertificatesJSON(proxyData.SslCertificatesJSON)
			}
			if proxyData.Region != "" {
				update.SetRegion(proxyData.Region)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update target HTTPS proxy %s: %w", proxyData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, proxyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for target HTTPS proxy %s: %w", proxyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, proxyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for target HTTPS proxy %s: %w", proxyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetHttpsProxies removes target HTTPS proxies not collected in the latest run.
func (s *Service) DeleteStaleTargetHttpsProxies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleProxies, err := tx.BronzeGCPComputeTargetHttpsProxy.Query().
		Where(
			bronzegcpcomputetargethttpsproxy.ProjectID(projectID),
			bronzegcpcomputetargethttpsproxy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, proxy := range staleProxies {
		if err := s.history.CloseHistory(ctx, tx, proxy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for target HTTPS proxy %s: %w", proxy.ID, err)
		}

		if err := tx.BronzeGCPComputeTargetHttpsProxy.DeleteOne(proxy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete target HTTPS proxy %s: %w", proxy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
