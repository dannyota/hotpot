package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpspannerdatabase"
)

// HistoryService manages Spanner database history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Spanner database.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DatabaseData, now time.Time) error {
	create := tx.BronzeHistoryGCPSpannerDatabase.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetState(data.State).
		SetCreateTime(data.CreateTime).
		SetVersionRetentionPeriod(data.VersionRetentionPeriod).
		SetEarliestVersionTime(data.EarliestVersionTime).
		SetDefaultLeader(data.DefaultLeader).
		SetDatabaseDialect(data.DatabaseDialect).
		SetEnableDropProtection(data.EnableDropProtection).
		SetReconciling(data.Reconciling).
		SetInstanceName(data.InstanceName).
		SetProjectID(data.ProjectID)

	if data.RestoreInfoJSON != nil {
		create.SetRestoreInfoJSON(data.RestoreInfoJSON)
	}
	if data.EncryptionConfigJSON != nil {
		create.SetEncryptionConfigJSON(data.EncryptionConfigJSON)
	}
	if data.EncryptionInfoJSON != nil {
		create.SetEncryptionInfoJSON(data.EncryptionInfoJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create Spanner database history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Spanner database.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSpannerDatabase, new *DatabaseData, diff *DatabaseDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSpannerDatabase.Query().
		Where(
			bronzehistorygcpspannerdatabase.ResourceID(old.ID),
			bronzehistorygcpspannerdatabase.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current Spanner database history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPSpannerDatabase.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current Spanner database history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPSpannerDatabase.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetState(new.State).
			SetCreateTime(new.CreateTime).
			SetVersionRetentionPeriod(new.VersionRetentionPeriod).
			SetEarliestVersionTime(new.EarliestVersionTime).
			SetDefaultLeader(new.DefaultLeader).
			SetDatabaseDialect(new.DatabaseDialect).
			SetEnableDropProtection(new.EnableDropProtection).
			SetReconciling(new.Reconciling).
			SetInstanceName(new.InstanceName).
			SetProjectID(new.ProjectID)

		if new.RestoreInfoJSON != nil {
			create.SetRestoreInfoJSON(new.RestoreInfoJSON)
		}
		if new.EncryptionConfigJSON != nil {
			create.SetEncryptionConfigJSON(new.EncryptionConfigJSON)
		}
		if new.EncryptionInfoJSON != nil {
			create.SetEncryptionInfoJSON(new.EncryptionInfoJSON)
		}

		_, err = create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new Spanner database history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Spanner database.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSpannerDatabase.Query().
		Where(
			bronzehistorygcpspannerdatabase.ResourceID(resourceID),
			bronzehistorygcpspannerdatabase.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current Spanner database history: %w", err)
	}

	err = tx.BronzeHistoryGCPSpannerDatabase.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close Spanner database history: %w", err)
	}

	return nil
}
