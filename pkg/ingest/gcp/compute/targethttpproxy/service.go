package targethttpproxy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputetargethttpproxy"
)

// Service handles GCP Compute target HTTP proxy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target HTTP proxy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target HTTP proxy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target HTTP proxy ingestion.
type IngestResult struct {
	ProjectID            string
	TargetHttpProxyCount int
	CollectedAt          time.Time
	DurationMillis       int64
}

// Ingest fetches target HTTP proxies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	proxies, err := s.client.ListTargetHttpProxies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target HTTP proxies: %w", err)
	}

	proxyDataList := make([]*TargetHttpProxyData, 0, len(proxies))
	for _, p := range proxies {
		data := ConvertTargetHttpProxy(p, params.ProjectID, collectedAt)
		proxyDataList = append(proxyDataList, data)
	}

	if err := s.saveTargetHttpProxies(ctx, proxyDataList); err != nil {
		return nil, fmt.Errorf("failed to save target HTTP proxies: %w", err)
	}

	return &IngestResult{
		ProjectID:            params.ProjectID,
		TargetHttpProxyCount: len(proxyDataList),
		CollectedAt:          collectedAt,
		DurationMillis:       time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveTargetHttpProxies(ctx context.Context, proxies []*TargetHttpProxyData) error {
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
		existing, err := tx.BronzeGCPComputeTargetHttpProxy.Query().
			Where(bronzegcpcomputetargethttpproxy.ID(proxyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing target HTTP proxy %s: %w", proxyData.Name, err)
		}

		diff := DiffTargetHttpProxyData(existing, proxyData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeTargetHttpProxy.UpdateOneID(proxyData.ID).
				SetCollectedAt(proxyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for target HTTP proxy %s: %w", proxyData.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeTargetHttpProxy.Create().
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
			if proxyData.ProxyBind {
				create.SetProxyBind(proxyData.ProxyBind)
			}
			if proxyData.HttpKeepAliveTimeoutSec != 0 {
				create.SetHTTPKeepAliveTimeoutSec(proxyData.HttpKeepAliveTimeoutSec)
			}
			if proxyData.Region != "" {
				create.SetRegion(proxyData.Region)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create target HTTP proxy %s: %w", proxyData.Name, err)
			}
		} else {
			update := tx.BronzeGCPComputeTargetHttpProxy.UpdateOneID(proxyData.ID).
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
			if proxyData.ProxyBind {
				update.SetProxyBind(proxyData.ProxyBind)
			}
			if proxyData.HttpKeepAliveTimeoutSec != 0 {
				update.SetHTTPKeepAliveTimeoutSec(proxyData.HttpKeepAliveTimeoutSec)
			}
			if proxyData.Region != "" {
				update.SetRegion(proxyData.Region)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update target HTTP proxy %s: %w", proxyData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, proxyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for target HTTP proxy %s: %w", proxyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, proxyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for target HTTP proxy %s: %w", proxyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetHttpProxies removes target HTTP proxies not collected in the latest run.
func (s *Service) DeleteStaleTargetHttpProxies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleProxies, err := tx.BronzeGCPComputeTargetHttpProxy.Query().
		Where(
			bronzegcpcomputetargethttpproxy.ProjectID(projectID),
			bronzegcpcomputetargethttpproxy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, proxy := range staleProxies {
		if err := s.history.CloseHistory(ctx, tx, proxy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for target HTTP proxy %s: %w", proxy.ID, err)
		}

		if err := tx.BronzeGCPComputeTargetHttpProxy.DeleteOne(proxy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete target HTTP proxy %s: %w", proxy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
