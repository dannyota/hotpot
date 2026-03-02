package installed_software

import (
	"context"
	"fmt"
	"time"

	entinventory "github.com/dannyota/hotpot/pkg/storage/ent/meec/inventory"
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
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes installed software that was not collected in the latest run.
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

	stale, err := tx.BronzeMEECInventoryInstalledSoftware.Query().
		Where(bronzemeecinventoryinstalledsoftware.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, sw := range stale {
		if err := s.history.CloseHistory(ctx, tx, sw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for installed software %s: %w", sw.ID, err)
		}

		if err := tx.BronzeMEECInventoryInstalledSoftware.DeleteOne(sw).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete installed software %s: %w", sw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
