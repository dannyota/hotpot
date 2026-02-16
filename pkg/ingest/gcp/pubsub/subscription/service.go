package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcppubsubsubscription"
)

// Service handles Pub/Sub subscription ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new subscription ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for subscription ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of subscription ingestion.
type IngestResult struct {
	ProjectID         string
	SubscriptionCount int
	CollectedAt       time.Time
	DurationMillis    int64
}

// Ingest fetches subscriptions from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch subscriptions from GCP
	rawSubscriptions, err := s.client.ListSubscriptions(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// Convert to subscription data
	subscriptionDataList := make([]*SubscriptionData, 0, len(rawSubscriptions))
	for _, raw := range rawSubscriptions {
		data, err := ConvertSubscription(raw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert subscription: %w", err)
		}
		subscriptionDataList = append(subscriptionDataList, data)
	}

	// Save to database
	if err := s.saveSubscriptions(ctx, subscriptionDataList); err != nil {
		return nil, fmt.Errorf("failed to save subscriptions: %w", err)
	}

	return &IngestResult{
		ProjectID:         params.ProjectID,
		SubscriptionCount: len(subscriptionDataList),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSubscriptions saves subscriptions to the database with history tracking.
func (s *Service) saveSubscriptions(ctx context.Context, subscriptions []*SubscriptionData) error {
	if len(subscriptions) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, subData := range subscriptions {
		// Load existing subscription
		existing, err := tx.BronzeGCPPubSubSubscription.Query().
			Where(bronzegcppubsubsubscription.ID(subData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing subscription %s: %w", subData.ID, err)
		}

		// Compute diff
		diff := DiffSubscriptionData(existing, subData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPPubSubSubscription.UpdateOneID(subData.ID).
				SetCollectedAt(subData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for subscription %s: %w", subData.ID, err)
			}
			continue
		}

		// Create or update subscription
		if existing == nil {
			create := tx.BronzeGCPPubSubSubscription.Create().
				SetID(subData.ID).
				SetName(subData.Name).
				SetProjectID(subData.ProjectID).
				SetCollectedAt(subData.CollectedAt).
				SetFirstCollectedAt(subData.CollectedAt).
				SetAckDeadlineSeconds(subData.AckDeadlineSeconds).
				SetRetainAckedMessages(subData.RetainAckedMessages).
				SetEnableMessageOrdering(subData.EnableMessageOrdering).
				SetDetached(subData.Detached).
				SetEnableExactlyOnceDelivery(subData.EnableExactlyOnceDelivery).
				SetState(subData.State)

			if subData.Topic != "" {
				create.SetTopic(subData.Topic)
			}
			if subData.Filter != "" {
				create.SetFilter(subData.Filter)
			}
			if subData.MessageRetentionDuration != "" {
				create.SetMessageRetentionDuration(subData.MessageRetentionDuration)
			}
			if subData.PushConfigJSON != nil {
				create.SetPushConfigJSON(subData.PushConfigJSON)
			}
			if subData.BigqueryConfigJSON != nil {
				create.SetBigqueryConfigJSON(subData.BigqueryConfigJSON)
			}
			if subData.CloudStorageConfigJSON != nil {
				create.SetCloudStorageConfigJSON(subData.CloudStorageConfigJSON)
			}
			if subData.LabelsJSON != nil {
				create.SetLabelsJSON(subData.LabelsJSON)
			}
			if subData.ExpirationPolicyJSON != nil {
				create.SetExpirationPolicyJSON(subData.ExpirationPolicyJSON)
			}
			if subData.DeadLetterPolicyJSON != nil {
				create.SetDeadLetterPolicyJSON(subData.DeadLetterPolicyJSON)
			}
			if subData.RetryPolicyJSON != nil {
				create.SetRetryPolicyJSON(subData.RetryPolicyJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create subscription %s: %w", subData.ID, err)
			}
		} else {
			update := tx.BronzeGCPPubSubSubscription.UpdateOneID(subData.ID).
				SetName(subData.Name).
				SetProjectID(subData.ProjectID).
				SetCollectedAt(subData.CollectedAt).
				SetAckDeadlineSeconds(subData.AckDeadlineSeconds).
				SetRetainAckedMessages(subData.RetainAckedMessages).
				SetEnableMessageOrdering(subData.EnableMessageOrdering).
				SetDetached(subData.Detached).
				SetEnableExactlyOnceDelivery(subData.EnableExactlyOnceDelivery).
				SetState(subData.State)

			if subData.Topic != "" {
				update.SetTopic(subData.Topic)
			}
			if subData.Filter != "" {
				update.SetFilter(subData.Filter)
			}
			if subData.MessageRetentionDuration != "" {
				update.SetMessageRetentionDuration(subData.MessageRetentionDuration)
			}
			if subData.PushConfigJSON != nil {
				update.SetPushConfigJSON(subData.PushConfigJSON)
			}
			if subData.BigqueryConfigJSON != nil {
				update.SetBigqueryConfigJSON(subData.BigqueryConfigJSON)
			}
			if subData.CloudStorageConfigJSON != nil {
				update.SetCloudStorageConfigJSON(subData.CloudStorageConfigJSON)
			}
			if subData.LabelsJSON != nil {
				update.SetLabelsJSON(subData.LabelsJSON)
			}
			if subData.ExpirationPolicyJSON != nil {
				update.SetExpirationPolicyJSON(subData.ExpirationPolicyJSON)
			}
			if subData.DeadLetterPolicyJSON != nil {
				update.SetDeadLetterPolicyJSON(subData.DeadLetterPolicyJSON)
			}
			if subData.RetryPolicyJSON != nil {
				update.SetRetryPolicyJSON(subData.RetryPolicyJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update subscription %s: %w", subData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, subData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for subscription %s: %w", subData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, subData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for subscription %s: %w", subData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSubscriptions removes subscriptions that were not collected in the latest run.
func (s *Service) DeleteStaleSubscriptions(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleSubs, err := tx.BronzeGCPPubSubSubscription.Query().
		Where(
			bronzegcppubsubsubscription.ProjectID(projectID),
			bronzegcppubsubsubscription.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, sub := range staleSubs {
		if err := s.history.CloseHistory(ctx, tx, sub.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for subscription %s: %w", sub.ID, err)
		}

		if err := tx.BronzeGCPPubSubSubscription.DeleteOne(sub).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete subscription %s: %w", sub.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
