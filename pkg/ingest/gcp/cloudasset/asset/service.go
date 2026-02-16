package asset

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcloudassetasset"
)

// Service handles Cloud Asset asset ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Cloud Asset asset ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Cloud Asset asset ingestion.
type IngestResult struct {
	AssetCount     int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Cloud Asset assets from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch assets from GCP
	rawAssets, err := s.client.ListAssets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}

	// Convert to asset data
	assetDataList := make([]*AssetData, 0, len(rawAssets))
	for _, raw := range rawAssets {
		data := ConvertAsset(raw.OrgName, raw.Asset, collectedAt)
		assetDataList = append(assetDataList, data)
	}

	// Save to database
	if err := s.saveAssets(ctx, assetDataList); err != nil {
		return nil, fmt.Errorf("failed to save assets: %w", err)
	}

	return &IngestResult{
		AssetCount:     len(assetDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveAssets saves Cloud Asset assets to the database with history tracking.
func (s *Service) saveAssets(ctx context.Context, assets []*AssetData) error {
	if len(assets) == 0 {
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

	for _, assetData := range assets {
		// Load existing asset
		existing, err := tx.BronzeGCPCloudAssetAsset.Query().
			Where(bronzegcpcloudassetasset.ID(assetData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing asset %s: %w", assetData.ID, err)
		}

		// Compute diff
		diff := DiffAssetData(existing, assetData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPCloudAssetAsset.UpdateOneID(assetData.ID).
				SetCollectedAt(assetData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for asset %s: %w", assetData.ID, err)
			}
			continue
		}

		// Create or update asset
		if existing == nil {
			create := tx.BronzeGCPCloudAssetAsset.Create().
				SetID(assetData.ID).
				SetAssetType(assetData.AssetType).
				SetOrganizationID(assetData.OrganizationID).
				SetCollectedAt(assetData.CollectedAt).
				SetFirstCollectedAt(assetData.CollectedAt)

			if assetData.UpdateTime != "" {
				create.SetUpdateTime(assetData.UpdateTime)
			}
			if assetData.ResourceJSON != nil {
				create.SetResourceJSON(assetData.ResourceJSON)
			}
			if assetData.IamPolicyJSON != nil {
				create.SetIamPolicyJSON(assetData.IamPolicyJSON)
			}
			if assetData.OrgPolicyJSON != nil {
				create.SetOrgPolicyJSON(assetData.OrgPolicyJSON)
			}
			if assetData.AccessPolicyJSON != nil {
				create.SetAccessPolicyJSON(assetData.AccessPolicyJSON)
			}
			if assetData.OsInventoryJSON != nil {
				create.SetOsInventoryJSON(assetData.OsInventoryJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create asset %s: %w", assetData.ID, err)
			}
		} else {
			update := tx.BronzeGCPCloudAssetAsset.UpdateOneID(assetData.ID).
				SetAssetType(assetData.AssetType).
				SetOrganizationID(assetData.OrganizationID).
				SetCollectedAt(assetData.CollectedAt)

			if assetData.UpdateTime != "" {
				update.SetUpdateTime(assetData.UpdateTime)
			}
			if assetData.ResourceJSON != nil {
				update.SetResourceJSON(assetData.ResourceJSON)
			}
			if assetData.IamPolicyJSON != nil {
				update.SetIamPolicyJSON(assetData.IamPolicyJSON)
			}
			if assetData.OrgPolicyJSON != nil {
				update.SetOrgPolicyJSON(assetData.OrgPolicyJSON)
			}
			if assetData.AccessPolicyJSON != nil {
				update.SetAccessPolicyJSON(assetData.AccessPolicyJSON)
			}
			if assetData.OsInventoryJSON != nil {
				update.SetOsInventoryJSON(assetData.OsInventoryJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update asset %s: %w", assetData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, assetData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for asset %s: %w", assetData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, assetData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for asset %s: %w", assetData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleAssets removes assets that were not collected in the latest run.
func (s *Service) DeleteStaleAssets(ctx context.Context, collectedAt time.Time) error {
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

	staleAssets, err := tx.BronzeGCPCloudAssetAsset.Query().
		Where(bronzegcpcloudassetasset.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, a := range staleAssets {
		if err := s.history.CloseHistory(ctx, tx, a.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for asset %s: %w", a.ID, err)
		}

		if err := tx.BronzeGCPCloudAssetAsset.DeleteOne(a).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete asset %s: %w", a.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
