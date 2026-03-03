package installed_software

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
	"github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory/bronzemeecinventorycomputer"
	"github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory/bronzemeecinventoryinstalledsoftware"
)

// Service handles MEEC installed software ingestion.
type Service struct {
	client    *Client
	entClient *entinventory.Client
	history   *HistoryService
}

// NewService creates a new installed software ingestion service.
func NewService(client *Client, entClient *entinventory.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// SaveComputerSoftware saves installed software for a single computer (upsert + history).
func (s *Service) SaveComputerSoftware(ctx context.Context, computerResourceID string, software []*InstalledSoftwareData) error {
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

	activeIDs := make(map[string]struct{}, len(software))

	for _, data := range software {
		existing, err := tx.BronzeMEECInventoryInstalledSoftware.Query().
			Where(bronzemeecinventoryinstalledsoftware.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !entinventory.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing installed software %s: %w", data.ResourceID, err)
		}

		diff := DiffInstalledSoftwareData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeMEECInventoryInstalledSoftware.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for installed software %s: %w", data.ResourceID, err)
			}
			activeIDs[data.ResourceID] = struct{}{}
			continue
		}

		if existing == nil {
			_, err := tx.BronzeMEECInventoryInstalledSoftware.Create().
				SetID(data.ResourceID).
				SetComputerResourceID(data.ComputerResourceID).
				SetSoftwareID(data.SoftwareID).
				SetSoftwareName(data.SoftwareName).
				SetSoftwareVersion(data.SoftwareVersion).
				SetDisplayName(data.DisplayName).
				SetManufacturerName(data.ManufacturerName).
				SetInstalledDate(data.InstalledDate).
				SetArchitecture(data.Architecture).
				SetLocation(data.Location).
				SetSwType(data.SwType).
				SetSwCategoryName(data.SwCategoryName).
				SetDetectedTime(data.DetectedTime).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create installed software %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for installed software %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeMEECInventoryInstalledSoftware.UpdateOneID(data.ResourceID).
				SetComputerResourceID(data.ComputerResourceID).
				SetSoftwareID(data.SoftwareID).
				SetSoftwareName(data.SoftwareName).
				SetSoftwareVersion(data.SoftwareVersion).
				SetDisplayName(data.DisplayName).
				SetManufacturerName(data.ManufacturerName).
				SetInstalledDate(data.InstalledDate).
				SetArchitecture(data.Architecture).
				SetLocation(data.Location).
				SetSwType(data.SwType).
				SetSwCategoryName(data.SwCategoryName).
				SetDetectedTime(data.DetectedTime).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update installed software %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for installed software %s: %w", data.ResourceID, err)
			}
		}

		activeIDs[data.ResourceID] = struct{}{}
	}

	// Delete stale installed software for this computer not returned by the API.
	dbSoftwareIDs, err := tx.BronzeMEECInventoryInstalledSoftware.Query().
		Where(bronzemeecinventoryinstalledsoftware.ComputerResourceIDEQ(computerResourceID)).
		Select(bronzemeecinventoryinstalledsoftware.FieldID).
		Strings(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query installed software IDs for computer %s: %w", computerResourceID, err)
	}

	staleCount := 0
	for _, id := range dbSoftwareIDs {
		if _, ok := activeIDs[id]; ok {
			continue
		}

		if err := s.history.CloseHistory(ctx, tx, id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for stale installed software %s: %w", id, err)
		}

		if err := tx.BronzeMEECInventoryInstalledSoftware.DeleteOneID(id).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete stale installed software %s: %w", id, err)
		}
		staleCount++
	}

	if staleCount > 0 {
		slog.Info("meec installed software: deleted stale for computer", "computerID", computerResourceID, "count", staleCount)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteOrphans removes installed software whose computer no longer exists.
func (s *Service) DeleteOrphans(ctx context.Context) error {
	now := time.Now()

	computerIDs, err := s.entClient.BronzeMEECInventoryComputer.Query().
		Select(bronzemeecinventorycomputer.FieldID).
		Strings(ctx)
	if err != nil {
		return fmt.Errorf("query computer IDs: %w", err)
	}

	orphans, err := s.entClient.BronzeMEECInventoryInstalledSoftware.Query().
		Where(bronzemeecinventoryinstalledsoftware.ComputerResourceIDNotIn(computerIDs...)).
		All(ctx)
	if err != nil {
		return fmt.Errorf("query orphan installed software: %w", err)
	}

	if len(orphans) == 0 {
		return nil
	}

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

	for _, sw := range orphans {
		if err := s.history.CloseHistory(ctx, tx, sw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for orphan installed software %s: %w", sw.ID, err)
		}

		if err := tx.BronzeMEECInventoryInstalledSoftware.DeleteOne(sw).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete orphan installed software %s: %w", sw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("meec installed software: deleted orphans", "count", len(orphans))

	return nil
}
