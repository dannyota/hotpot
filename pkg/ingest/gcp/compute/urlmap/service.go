package urlmap

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeurlmap"
)

// Service handles GCP Compute URL map ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new URL map ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for URL map ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of URL map ingestion.
type IngestResult struct {
	ProjectID      string
	UrlMapCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches URL maps from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	urlMaps, err := s.client.ListUrlMaps(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list URL maps: %w", err)
	}

	dataList := make([]*UrlMapData, 0, len(urlMaps))
	for _, um := range urlMaps {
		data, err := ConvertUrlMap(um, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert URL map: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := s.saveUrlMaps(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save URL maps: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		UrlMapCount:    len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveUrlMaps(ctx context.Context, urlMaps []*UrlMapData) error {
	if len(urlMaps) == 0 {
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

	for _, umData := range urlMaps {
		existing, err := tx.BronzeGCPComputeUrlMap.Query().
			Where(bronzegcpcomputeurlmap.ID(umData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing URL map %s: %w", umData.Name, err)
		}

		diff := DiffUrlMapData(existing, umData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeUrlMap.UpdateOneID(umData.ID).
				SetCollectedAt(umData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for URL map %s: %w", umData.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeUrlMap.Create().
				SetID(umData.ID).
				SetName(umData.Name).
				SetProjectID(umData.ProjectID).
				SetCollectedAt(umData.CollectedAt).
				SetFirstCollectedAt(umData.CollectedAt)

			if umData.Description != "" {
				create.SetDescription(umData.Description)
			}
			if umData.CreationTimestamp != "" {
				create.SetCreationTimestamp(umData.CreationTimestamp)
			}
			if umData.SelfLink != "" {
				create.SetSelfLink(umData.SelfLink)
			}
			if umData.Fingerprint != "" {
				create.SetFingerprint(umData.Fingerprint)
			}
			if umData.DefaultService != "" {
				create.SetDefaultService(umData.DefaultService)
			}
			if umData.Region != "" {
				create.SetRegion(umData.Region)
			}
			if umData.HostRulesJSON != nil {
				create.SetHostRulesJSON(umData.HostRulesJSON)
			}
			if umData.PathMatchersJSON != nil {
				create.SetPathMatchersJSON(umData.PathMatchersJSON)
			}
			if umData.TestsJSON != nil {
				create.SetTestsJSON(umData.TestsJSON)
			}
			if umData.DefaultRouteActionJSON != nil {
				create.SetDefaultRouteActionJSON(umData.DefaultRouteActionJSON)
			}
			if umData.DefaultUrlRedirectJSON != nil {
				create.SetDefaultURLRedirectJSON(umData.DefaultUrlRedirectJSON)
			}
			if umData.HeaderActionJSON != nil {
				create.SetHeaderActionJSON(umData.HeaderActionJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create URL map %s: %w", umData.Name, err)
			}
		} else {
			update := tx.BronzeGCPComputeUrlMap.UpdateOneID(umData.ID).
				SetName(umData.Name).
				SetProjectID(umData.ProjectID).
				SetCollectedAt(umData.CollectedAt)

			if umData.Description != "" {
				update.SetDescription(umData.Description)
			}
			if umData.CreationTimestamp != "" {
				update.SetCreationTimestamp(umData.CreationTimestamp)
			}
			if umData.SelfLink != "" {
				update.SetSelfLink(umData.SelfLink)
			}
			if umData.Fingerprint != "" {
				update.SetFingerprint(umData.Fingerprint)
			}
			if umData.DefaultService != "" {
				update.SetDefaultService(umData.DefaultService)
			}
			if umData.Region != "" {
				update.SetRegion(umData.Region)
			}
			if umData.HostRulesJSON != nil {
				update.SetHostRulesJSON(umData.HostRulesJSON)
			}
			if umData.PathMatchersJSON != nil {
				update.SetPathMatchersJSON(umData.PathMatchersJSON)
			}
			if umData.TestsJSON != nil {
				update.SetTestsJSON(umData.TestsJSON)
			}
			if umData.DefaultRouteActionJSON != nil {
				update.SetDefaultRouteActionJSON(umData.DefaultRouteActionJSON)
			}
			if umData.DefaultUrlRedirectJSON != nil {
				update.SetDefaultURLRedirectJSON(umData.DefaultUrlRedirectJSON)
			}
			if umData.HeaderActionJSON != nil {
				update.SetHeaderActionJSON(umData.HeaderActionJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update URL map %s: %w", umData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, umData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for URL map %s: %w", umData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, umData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for URL map %s: %w", umData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleUrlMaps removes URL maps not collected in the latest run.
func (s *Service) DeleteStaleUrlMaps(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPComputeUrlMap.Query().
		Where(
			bronzegcpcomputeurlmap.ProjectID(projectID),
			bronzegcpcomputeurlmap.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, um := range stale {
		if err := s.history.CloseHistory(ctx, tx, um.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for URL map %s: %w", um.ID, err)
		}

		if err := tx.BronzeGCPComputeUrlMap.DeleteOne(um).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete URL map %s: %w", um.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
