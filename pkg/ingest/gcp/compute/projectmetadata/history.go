package projectmetadata

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeprojectmetadata"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeprojectmetadataitem"
)

// HistoryService handles history tracking for project metadata.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for new project metadata and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ProjectMetadataData, now time.Time) error {
	// Create metadata history
	metaHist, err := tx.BronzeHistoryGCPComputeProjectMetadata.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDefaultServiceAccount(data.DefaultServiceAccount).
		SetDefaultNetworkTier(data.DefaultNetworkTier).
		SetXpnProjectStatus(data.XpnProjectStatus).
		SetCreationTimestamp(data.CreationTimestamp).
		SetUsageExportLocationJSON(data.UsageExportLocationJSON).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create project metadata history: %w", err)
	}

	// Create children history with metadata_history_id
	return h.createChildrenHistory(ctx, tx, metaHist.HistoryID, data, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeProjectMetadata, new *ProjectMetadataData, diff *ProjectMetadataDiff, now time.Time) error {
	// Get current metadata history
	currentHist, err := tx.BronzeHistoryGCPComputeProjectMetadata.Query().
		Where(
			bronzehistorygcpcomputeprojectmetadata.ResourceID(old.ID),
			bronzehistorygcpcomputeprojectmetadata.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current project metadata history: %w", err)
	}

	// If metadata-level fields changed, close old and create new metadata history
	if diff.IsChanged {
		// Close old metadata history
		err = tx.BronzeHistoryGCPComputeProjectMetadata.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current project metadata history: %w", err)
		}

		// Create new metadata history
		metaHist, err := tx.BronzeHistoryGCPComputeProjectMetadata.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDefaultServiceAccount(new.DefaultServiceAccount).
			SetDefaultNetworkTier(new.DefaultNetworkTier).
			SetXpnProjectStatus(new.XpnProjectStatus).
			SetCreationTimestamp(new.CreationTimestamp).
			SetUsageExportLocationJSON(new.UsageExportLocationJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new project metadata history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, metaHist.HistoryID, new, now)
	}

	// Metadata unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for deleted project metadata.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current metadata history
	currentHist, err := tx.BronzeHistoryGCPComputeProjectMetadata.Query().
		Where(
			bronzehistorygcpcomputeprojectmetadata.ResourceID(resourceID),
			bronzehistorygcpcomputeprojectmetadata.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current project metadata history: %w", err)
	}

	// Close metadata history
	err = tx.BronzeHistoryGCPComputeProjectMetadata.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close project metadata history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, metadataHistoryID uint, data *ProjectMetadataData, now time.Time) error {
	for _, item := range data.Items {
		create := tx.BronzeHistoryGCPComputeProjectMetadataItem.Create().
			SetMetadataHistoryID(metadataHistoryID).
			SetValidFrom(now).
			SetKey(item.Key)
		if item.Value != "" {
			create.SetValue(item.Value)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create item history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, metadataHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeProjectMetadataItem.Update().
		Where(
			bronzehistorygcpcomputeprojectmetadataitem.MetadataHistoryID(metadataHistoryID),
			bronzehistorygcpcomputeprojectmetadataitem.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close item history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, metadataHistoryID uint, new *ProjectMetadataData, diff *ProjectMetadataDiff, now time.Time) error {
	if diff.ItemsDiff.Changed {
		if err := h.updateItemsHistory(ctx, tx, metadataHistoryID, new.Items, now); err != nil {
			return fmt.Errorf("failed to update items history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateItemsHistory(ctx context.Context, tx *ent.Tx, metadataHistoryID uint, items []ItemData, now time.Time) error {
	// Close old items history
	_, err := tx.BronzeHistoryGCPComputeProjectMetadataItem.Update().
		Where(
			bronzehistorygcpcomputeprojectmetadataitem.MetadataHistoryID(metadataHistoryID),
			bronzehistorygcpcomputeprojectmetadataitem.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close items history: %w", err)
	}

	// Create new items history
	for _, item := range items {
		create := tx.BronzeHistoryGCPComputeProjectMetadataItem.Create().
			SetMetadataHistoryID(metadataHistoryID).
			SetValidFrom(now).
			SetKey(item.Key)
		if item.Value != "" {
			create.SetValue(item.Value)
		}
		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create item history: %w", err)
		}
	}

	return nil
}
