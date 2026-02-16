package table

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbigquerytable"
)

// HistoryService manages BigQuery table history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new BigQuery table.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TableData, now time.Time) error {
	create := tx.BronzeHistoryGCPBigQueryTable.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDatasetID(data.DatasetID).
		SetRequirePartitionFilter(data.RequirePartitionFilter).
		SetProjectID(data.ProjectID)

	if data.FriendlyName != "" {
		create.SetFriendlyName(data.FriendlyName)
	}
	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.SchemaJSON != nil {
		create.SetSchemaJSON(data.SchemaJSON)
	}
	if data.NumBytes != nil {
		create.SetNumBytes(*data.NumBytes)
	}
	if data.NumLongTermBytes != nil {
		create.SetNumLongTermBytes(*data.NumLongTermBytes)
	}
	if data.NumRows != nil {
		create.SetNumRows(*data.NumRows)
	}
	if data.CreationTime != "" {
		create.SetCreationTime(data.CreationTime)
	}
	if data.ExpirationTime != "" {
		create.SetExpirationTime(data.ExpirationTime)
	}
	if data.LastModifiedTime != "" {
		create.SetLastModifiedTime(data.LastModifiedTime)
	}
	if data.TableType != "" {
		create.SetTableType(data.TableType)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.EncryptionConfigurationJSON != nil {
		create.SetEncryptionConfigurationJSON(data.EncryptionConfigurationJSON)
	}
	if data.TimePartitioningJSON != nil {
		create.SetTimePartitioningJSON(data.TimePartitioningJSON)
	}
	if data.RangePartitioningJSON != nil {
		create.SetRangePartitioningJSON(data.RangePartitioningJSON)
	}
	if data.ClusteringJSON != nil {
		create.SetClusteringJSON(data.ClusteringJSON)
	}
	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create BigQuery table history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed BigQuery table.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBigQueryTable, new *TableData, diff *TableDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigQueryTable.Query().
		Where(
			bronzehistorygcpbigquerytable.ResourceID(old.ID),
			bronzehistorygcpbigquerytable.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current BigQuery table history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBigQueryTable.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current BigQuery table history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPBigQueryTable.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDatasetID(new.DatasetID).
			SetRequirePartitionFilter(new.RequirePartitionFilter).
			SetProjectID(new.ProjectID)

		if new.FriendlyName != "" {
			create.SetFriendlyName(new.FriendlyName)
		}
		if new.Description != "" {
			create.SetDescription(new.Description)
		}
		if new.SchemaJSON != nil {
			create.SetSchemaJSON(new.SchemaJSON)
		}
		if new.NumBytes != nil {
			create.SetNumBytes(*new.NumBytes)
		}
		if new.NumLongTermBytes != nil {
			create.SetNumLongTermBytes(*new.NumLongTermBytes)
		}
		if new.NumRows != nil {
			create.SetNumRows(*new.NumRows)
		}
		if new.CreationTime != "" {
			create.SetCreationTime(new.CreationTime)
		}
		if new.ExpirationTime != "" {
			create.SetExpirationTime(new.ExpirationTime)
		}
		if new.LastModifiedTime != "" {
			create.SetLastModifiedTime(new.LastModifiedTime)
		}
		if new.TableType != "" {
			create.SetTableType(new.TableType)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.EncryptionConfigurationJSON != nil {
			create.SetEncryptionConfigurationJSON(new.EncryptionConfigurationJSON)
		}
		if new.TimePartitioningJSON != nil {
			create.SetTimePartitioningJSON(new.TimePartitioningJSON)
		}
		if new.RangePartitioningJSON != nil {
			create.SetRangePartitioningJSON(new.RangePartitioningJSON)
		}
		if new.ClusteringJSON != nil {
			create.SetClusteringJSON(new.ClusteringJSON)
		}
		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new BigQuery table history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted BigQuery table.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBigQueryTable.Query().
		Where(
			bronzehistorygcpbigquerytable.ResourceID(resourceID),
			bronzehistorygcpbigquerytable.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current BigQuery table history: %w", err)
	}

	err = tx.BronzeHistoryGCPBigQueryTable.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close BigQuery table history: %w", err)
	}

	return nil
}
