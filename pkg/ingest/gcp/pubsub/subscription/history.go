package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcppubsubsubscription"
)

// HistoryService manages Pub/Sub subscription history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new subscription.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SubscriptionData, now time.Time) error {
	create := tx.BronzeHistoryGCPPubSubSubscription.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetAckDeadlineSeconds(data.AckDeadlineSeconds).
		SetRetainAckedMessages(data.RetainAckedMessages).
		SetEnableMessageOrdering(data.EnableMessageOrdering).
		SetDetached(data.Detached).
		SetEnableExactlyOnceDelivery(data.EnableExactlyOnceDelivery).
		SetState(data.State).
		SetProjectID(data.ProjectID)

	if data.Topic != "" {
		create.SetTopic(data.Topic)
	}
	if data.Filter != "" {
		create.SetFilter(data.Filter)
	}
	if data.MessageRetentionDuration != "" {
		create.SetMessageRetentionDuration(data.MessageRetentionDuration)
	}
	if data.PushConfigJSON != nil {
		create.SetPushConfigJSON(data.PushConfigJSON)
	}
	if data.BigqueryConfigJSON != nil {
		create.SetBigqueryConfigJSON(data.BigqueryConfigJSON)
	}
	if data.CloudStorageConfigJSON != nil {
		create.SetCloudStorageConfigJSON(data.CloudStorageConfigJSON)
	}
	if data.LabelsJSON != nil {
		create.SetLabelsJSON(data.LabelsJSON)
	}
	if data.ExpirationPolicyJSON != nil {
		create.SetExpirationPolicyJSON(data.ExpirationPolicyJSON)
	}
	if data.DeadLetterPolicyJSON != nil {
		create.SetDeadLetterPolicyJSON(data.DeadLetterPolicyJSON)
	}
	if data.RetryPolicyJSON != nil {
		create.SetRetryPolicyJSON(data.RetryPolicyJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create subscription history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed subscription.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPPubSubSubscription, new *SubscriptionData, diff *SubscriptionDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPPubSubSubscription.Query().
		Where(
			bronzehistorygcppubsubsubscription.ResourceID(old.ID),
			bronzehistorygcppubsubsubscription.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current subscription history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPPubSubSubscription.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current subscription history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPPubSubSubscription.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetAckDeadlineSeconds(new.AckDeadlineSeconds).
			SetRetainAckedMessages(new.RetainAckedMessages).
			SetEnableMessageOrdering(new.EnableMessageOrdering).
			SetDetached(new.Detached).
			SetEnableExactlyOnceDelivery(new.EnableExactlyOnceDelivery).
			SetState(new.State).
			SetProjectID(new.ProjectID)

		if new.Topic != "" {
			create.SetTopic(new.Topic)
		}
		if new.Filter != "" {
			create.SetFilter(new.Filter)
		}
		if new.MessageRetentionDuration != "" {
			create.SetMessageRetentionDuration(new.MessageRetentionDuration)
		}
		if new.PushConfigJSON != nil {
			create.SetPushConfigJSON(new.PushConfigJSON)
		}
		if new.BigqueryConfigJSON != nil {
			create.SetBigqueryConfigJSON(new.BigqueryConfigJSON)
		}
		if new.CloudStorageConfigJSON != nil {
			create.SetCloudStorageConfigJSON(new.CloudStorageConfigJSON)
		}
		if new.LabelsJSON != nil {
			create.SetLabelsJSON(new.LabelsJSON)
		}
		if new.ExpirationPolicyJSON != nil {
			create.SetExpirationPolicyJSON(new.ExpirationPolicyJSON)
		}
		if new.DeadLetterPolicyJSON != nil {
			create.SetDeadLetterPolicyJSON(new.DeadLetterPolicyJSON)
		}
		if new.RetryPolicyJSON != nil {
			create.SetRetryPolicyJSON(new.RetryPolicyJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new subscription history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted subscription.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPPubSubSubscription.Query().
		Where(
			bronzehistorygcppubsubsubscription.ResourceID(resourceID),
			bronzehistorygcppubsubsubscription.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current subscription history: %w", err)
	}

	err = tx.BronzeHistoryGCPPubSubSubscription.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close subscription history: %w", err)
	}

	return nil
}
