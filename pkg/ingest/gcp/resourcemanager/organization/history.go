package organization

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcporganization"
)

// HistoryService manages organization history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new organization.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, orgData *OrganizationData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPOrganization.Create().
		SetResourceID(orgData.ID).
		SetValidFrom(now).
		SetCollectedAt(orgData.CollectedAt).
		SetFirstCollectedAt(orgData.CollectedAt).
		SetName(orgData.Name).
		SetDisplayName(orgData.DisplayName).
		SetState(orgData.State).
		SetDirectoryCustomerID(orgData.DirectoryCustomerID).
		SetEtag(orgData.Etag).
		SetCreateTime(orgData.CreateTime).
		SetUpdateTime(orgData.UpdateTime).
		SetDeleteTime(orgData.DeleteTime).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create organization history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed organization.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPOrganization, new *OrganizationData, diff *OrganizationDiff, now time.Time) error {
	// Get current organization history
	currentHistory, err := tx.BronzeHistoryGCPOrganization.Query().
		Where(
			bronzehistorygcporganization.ResourceID(old.ID),
			bronzehistorygcporganization.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current organization history: %w", err)
	}

	// Close current organization history if core fields changed
	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPOrganization.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current organization history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPOrganization.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDisplayName(new.DisplayName).
			SetState(new.State).
			SetDirectoryCustomerID(new.DirectoryCustomerID).
			SetEtag(new.Etag).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
			SetDeleteTime(new.DeleteTime).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new organization history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted organization.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, organizationID string, now time.Time) error {
	// Get current organization history
	currentHistory, err := tx.BronzeHistoryGCPOrganization.Query().
		Where(
			bronzehistorygcporganization.ResourceID(organizationID),
			bronzehistorygcporganization.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current organization history: %w", err)
	}

	// Close organization history
	err = tx.BronzeHistoryGCPOrganization.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close organization history: %w", err)
	}

	return nil
}
