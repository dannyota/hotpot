package site

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzes1site"
)

// Service handles SentinelOne site ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new site ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of site ingestion.
type IngestResult struct {
	SiteCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all sites from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allSites []*SiteData
	cursor := ""

	for {
		batch, err := s.client.GetSitesBatch(cursor)
		if err != nil {
			return nil, fmt.Errorf("get sites batch: %w", err)
		}

		for _, apiSite := range batch.Sites {
			allSites = append(allSites, ConvertSite(apiSite, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveSites(ctx, allSites); err != nil {
		return nil, fmt.Errorf("save sites: %w", err)
	}

	return &IngestResult{
		SiteCount:      len(allSites),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSites(ctx context.Context, sites []*SiteData) error {
	if len(sites) == 0 {
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

	for _, data := range sites {
		existing, err := tx.BronzeS1Site.Query().
			Where(bronzes1site.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing site %s: %w", data.ResourceID, err)
		}

		diff := DiffSiteData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1Site.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for site %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1Site.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetState(data.State).
				SetSiteType(data.SiteType).
				SetSuite(data.Suite).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetHealthStatus(data.HealthStatus).
				SetActiveLicenses(data.ActiveLicenses).
				SetTotalLicenses(data.TotalLicenses).
				SetUnlimitedLicenses(data.UnlimitedLicenses).
				SetIsDefault(data.IsDefault).
				SetDescription(data.Description).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.Expiration != nil {
				create.SetExpiration(*data.Expiration)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create site %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for site %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1Site.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetAccountID(data.AccountID).
				SetAccountName(data.AccountName).
				SetState(data.State).
				SetSiteType(data.SiteType).
				SetSuite(data.Suite).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetHealthStatus(data.HealthStatus).
				SetActiveLicenses(data.ActiveLicenses).
				SetTotalLicenses(data.TotalLicenses).
				SetUnlimitedLicenses(data.UnlimitedLicenses).
				SetIsDefault(data.IsDefault).
				SetDescription(data.Description).
				SetCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.Expiration != nil {
				update.SetExpiration(*data.Expiration)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update site %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for site %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes sites that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeS1Site.Query().
		Where(bronzes1site.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, s1Site := range stale {
		if err := s.history.CloseHistory(ctx, tx, s1Site.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for site %s: %w", s1Site.ID, err)
		}

		if err := tx.BronzeS1Site.DeleteOne(s1Site).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete site %s: %w", s1Site.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
