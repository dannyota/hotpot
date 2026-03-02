package software

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
	"github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory/bronzemeecinventorysoftware"
)

// Service handles MEEC software ingestion.
type Service struct {
	client    *Client
	entClient *entinventory.Client
	history   *HistoryService
}

// NewService creates a new software ingestion service.
func NewService(client *Client, entClient *entinventory.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of software ingestion.
type IngestResult struct {
	SoftwareCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all software entries from MEEC using page-based pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allSoftware []*SoftwareData
	page := 1

	for {
		batch, err := s.client.GetSoftwareBatch(page)
		if err != nil {
			slog.Error("meec software batch failed", "page", page, "totalSoFar", len(allSoftware), "error", err)
			return nil, fmt.Errorf("get software batch: %w", err)
		}

		for _, sw := range batch.Software {
			allSoftware = append(allSoftware, ConvertSoftware(sw, collectedAt))
		}

		slog.Info("meec software batch fetched", "page", page, "batchItems", len(batch.Software), "totalFetched", len(allSoftware), "totalExpected", batch.Total, "hasMore", batch.HasMore)

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		page++
	}

	if err := s.saveSoftware(ctx, allSoftware); err != nil {
		return nil, fmt.Errorf("save software: %w", err)
	}

	return &IngestResult{
		SoftwareCount:  len(allSoftware),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSoftware(ctx context.Context, software []*SoftwareData) error {
	if len(software) == 0 {
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

	for _, data := range software {
		existing, err := tx.BronzeMEECInventorySoftware.Query().
			Where(bronzemeecinventorysoftware.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !entinventory.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing software %s: %w", data.ResourceID, err)
		}

		diff := DiffSoftwareData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeMEECInventorySoftware.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for software %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			_, err := tx.BronzeMEECInventorySoftware.Create().
				SetID(data.ResourceID).
				SetSoftwareName(data.SoftwareName).
				SetSoftwareVersion(data.SoftwareVersion).
				SetDisplayName(data.DisplayName).
				SetManufacturerID(data.ManufacturerID).
				SetManufacturerName(data.ManufacturerName).
				SetSwCategoryName(data.SwCategoryName).
				SetSwType(data.SwType).
				SetSwFamily(data.SwFamily).
				SetInstalledFormat(data.InstalledFormat).
				SetIsUsageProhibited(data.IsUsageProhibited).
				SetManagedInstallations(data.ManagedInstallations).
				SetNetworkInstallations(data.NetworkInstallations).
				SetManagedSwID(data.ManagedSwID).
				SetDetectedTime(data.DetectedTime).
				SetCompliantStatus(data.CompliantStatus).
				SetTotalCopies(data.TotalCopies).
				SetRemainingCopies(data.RemainingCopies).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create software %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for software %s: %w", data.ResourceID, err)
			}
		} else {
			_, err := tx.BronzeMEECInventorySoftware.UpdateOneID(data.ResourceID).
				SetSoftwareName(data.SoftwareName).
				SetSoftwareVersion(data.SoftwareVersion).
				SetDisplayName(data.DisplayName).
				SetManufacturerID(data.ManufacturerID).
				SetManufacturerName(data.ManufacturerName).
				SetSwCategoryName(data.SwCategoryName).
				SetSwType(data.SwType).
				SetSwFamily(data.SwFamily).
				SetInstalledFormat(data.InstalledFormat).
				SetIsUsageProhibited(data.IsUsageProhibited).
				SetManagedInstallations(data.ManagedInstallations).
				SetNetworkInstallations(data.NetworkInstallations).
				SetManagedSwID(data.ManagedSwID).
				SetDetectedTime(data.DetectedTime).
				SetCompliantStatus(data.CompliantStatus).
				SetTotalCopies(data.TotalCopies).
				SetRemainingCopies(data.RemainingCopies).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update software %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for software %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes software entries that were not collected in the latest run.
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

	stale, err := tx.BronzeMEECInventorySoftware.Query().
		Where(bronzemeecinventorysoftware.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, sw := range stale {
		if err := s.history.CloseHistory(ctx, tx, sw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for software %s: %w", sw.ID, err)
		}

		if err := tx.BronzeMEECInventorySoftware.DeleteOne(sw).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete software %s: %w", sw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
