package dataset

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbigquerydataset"
)

// HistoryService manages BigQuery dataset history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new BigQuery dataset.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *DatasetData, now time.Time) error {
	create := tx.BronzeHistoryGCPBigQueryDataset.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetProjectID(data.ProjectID)

	if data.FriendlyName != "" {
		create.SetFriendlyName(data.FriendlyName)
	}
	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.Location != "" {
		create.SetLocation(data.Location)
	}
	if data.DefaultTableExpirationMs != nil {
		create.SetDefaultTableExpirationMs(*data.DefaultTableExpirationMs)
	}
	if data.DefaultPartitionExpirationMs != nil {
		create.SetDefaultPartitionExpirationMs(*data.DefaultPartitionExpirationMs)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.AccessJSON != nil {
		create.SetAccessJSON(data.AccessJSON)
	}
	if data.CreationTime != "" {
		create.SetCreationTime(data.CreationTime)
	}
	if data.LastModifiedTime != "" {
		create.SetLastModifiedTime(data.LastModifiedTime)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}
	if data.DefaultCollation != "" {
		create.SetDefaultCollation(data.DefaultCollation)
	}
	if data.MaxTimeTravelHours != nil {
		create.SetMaxTimeTravelHours(*data.MaxTimeTravelHours)
	}
	if data.DefaultEncryptionConfigurationJSON != nil {
		create.SetDefaultEncryptionConfigurationJSON(data.DefaultEncryptionConfigurationJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create BigQuery dataset history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed BigQuery dataset.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBigQueryDataset, new *DatasetData, diff *DatasetDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigQueryDataset.Query().
		Where(
			bronzehistorygcpbigquerydataset.ResourceID(old.ID),
			bronzehistorygcpbigquerydataset.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current BigQuery dataset history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBigQueryDataset.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current BigQuery dataset history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPBigQueryDataset.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetProjectID(new.ProjectID)

		if new.FriendlyName != "" {
			create.SetFriendlyName(new.FriendlyName)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.Location != "" {
			create.SetLocation(new.Location)
		}
		if new.DefaultTableExpirationMs != nil {
			create.SetDefaultTableExpirationMs(*new.DefaultTableExpirationMs)
		}
		if new.DefaultPartitionExpirationMs != nil {
			create.SetDefaultPartitionExpirationMs(*new.DefaultPartitionExpirationMs)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.AccessJSON != nil {
			create.SetAccessJSON(new.AccessJSON)
		}
		if new.CreationTime != "" {
			create.SetCreationTime(new.CreationTime)
		}
		if new.LastModifiedTime != "" {
			create.SetLastModifiedTime(new.LastModifiedTime)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}
		if new.DefaultCollation != "" {
			create.SetDefaultCollation(new.DefaultCollation)
		}
		if new.MaxTimeTravelHours != nil {
			create.SetMaxTimeTravelHours(*new.MaxTimeTravelHours)
		}
		if new.DefaultEncryptionConfigurationJSON != nil {
			create.SetDefaultEncryptionConfigurationJSON(new.DefaultEncryptionConfigurationJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new BigQuery dataset history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted BigQuery dataset.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigQueryDataset.Query().
		Where(
			bronzehistorygcpbigquerydataset.ResourceID(resourceID),
			bronzehistorygcpbigquerydataset.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current BigQuery dataset history: %w", err)
	}

	err = tx.BronzeHistoryGCPBigQueryDataset.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close BigQuery dataset history: %w", err)
	}

	return nil
}
