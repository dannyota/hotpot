package notificationconfig

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsecuritycenternotificationconfig"
)

// HistoryService manages SCC notification config history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new SCC notification config.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *NotificationConfigData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPSecurityCenterNotificationConfig.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetPubsubTopic(data.PubsubTopic).
		SetStreamingConfigJSON(data.StreamingConfigJSON).
		SetServiceAccount(data.ServiceAccount).
		SetOrganizationID(data.OrganizationID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create SCC notification config history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed SCC notification config.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSecurityCenterNotificationConfig, new *NotificationConfigData, diff *NotificationConfigDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterNotificationConfig.Query().
		Where(
			bronzehistorygcpsecuritycenternotificationconfig.ResourceID(old.ID),
			bronzehistorygcpsecuritycenternotificationconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current SCC notification config history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPSecurityCenterNotificationConfig.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current SCC notification config history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPSecurityCenterNotificationConfig.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetPubsubTopic(new.PubsubTopic).
			SetStreamingConfigJSON(new.StreamingConfigJSON).
			SetServiceAccount(new.ServiceAccount).
			SetOrganizationID(new.OrganizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new SCC notification config history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted SCC notification config.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterNotificationConfig.Query().
		Where(
			bronzehistorygcpsecuritycenternotificationconfig.ResourceID(resourceID),
			bronzehistorygcpsecuritycenternotificationconfig.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current SCC notification config history: %w", err)
	}

	err = tx.BronzeHistoryGCPSecurityCenterNotificationConfig.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close SCC notification config history: %w", err)
	}

	return nil
}
