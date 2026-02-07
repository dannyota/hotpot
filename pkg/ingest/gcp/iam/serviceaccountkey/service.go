package serviceaccountkey

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

type IngestParams struct {
	ProjectID string
}

type IngestResult struct {
	ProjectID              string
	ServiceAccountKeyCount int
	CollectedAt            time.Time
	DurationMillis         int64
}

func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	keysWithAccounts, err := s.client.ListServiceAccountKeys(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service account keys: %w", err)
	}

	bronzeKeys := make([]bronze.GCPIAMServiceAccountKey, 0, len(keysWithAccounts))
	for _, kwa := range keysWithAccounts {
		bronzeKeys = append(bronzeKeys, ConvertServiceAccountKey(kwa, params.ProjectID, collectedAt))
	}

	if err := s.saveServiceAccountKeys(ctx, bronzeKeys); err != nil {
		return nil, fmt.Errorf("failed to save service account keys: %w", err)
	}

	return &IngestResult{
		ProjectID:              params.ProjectID,
		ServiceAccountKeyCount: len(bronzeKeys),
		CollectedAt:            collectedAt,
		DurationMillis:         time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveServiceAccountKeys(ctx context.Context, keys []bronze.GCPIAMServiceAccountKey) error {
	if len(keys) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, key := range keys {
			var existing *bronze.GCPIAMServiceAccountKey
			var old bronze.GCPIAMServiceAccountKey
			err := tx.Where("resource_id = ?", key.ResourceID).First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing key %s: %w", key.ResourceID, err)
			}

			diff := DiffServiceAccountKey(existing, &key)

			if !diff.HasAnyChange() && existing != nil {
				if err := tx.Model(&bronze.GCPIAMServiceAccountKey{}).
					Where("resource_id = ?", key.ResourceID).
					Update("collected_at", key.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for key %s: %w", key.ResourceID, err)
				}
				continue
			}

			if err := tx.Save(&key).Error; err != nil {
				return fmt.Errorf("failed to upsert key %s: %w", key.ResourceID, err)
			}

			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &key, now); err != nil {
					return fmt.Errorf("failed to create history for key %s: %w", key.ResourceID, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &key, diff, now); err != nil {
					return fmt.Errorf("failed to update history for key %s: %w", key.ResourceID, err)
				}
			}
		}
		return nil
	})
}

func (s *Service) DeleteStaleKeys(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stale []bronze.GCPIAMServiceAccountKey
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&stale).Error; err != nil {
			return err
		}

		for _, key := range stale {
			if err := s.history.CloseHistory(tx, key.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for key %s: %w", key.ResourceID, err)
			}
			if err := tx.Delete(&key).Error; err != nil {
				return fmt.Errorf("failed to delete key %s: %w", key.ResourceID, err)
			}
		}
		return nil
	})
}
