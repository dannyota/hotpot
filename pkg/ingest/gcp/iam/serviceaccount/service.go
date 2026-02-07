package serviceaccount

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
	ProjectID           string
	ServiceAccountCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	accounts, err := s.client.ListServiceAccounts(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}

	bronzeAccounts := make([]bronze.GCPIAMServiceAccount, 0, len(accounts))
	for _, sa := range accounts {
		bronzeAccounts = append(bronzeAccounts, ConvertServiceAccount(sa, params.ProjectID, collectedAt))
	}

	if err := s.saveServiceAccounts(ctx, bronzeAccounts); err != nil {
		return nil, fmt.Errorf("failed to save service accounts: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		ServiceAccountCount: len(bronzeAccounts),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveServiceAccounts(ctx context.Context, accounts []bronze.GCPIAMServiceAccount) error {
	if len(accounts) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sa := range accounts {
			var existing *bronze.GCPIAMServiceAccount
			var old bronze.GCPIAMServiceAccount
			err := tx.Where("resource_id = ?", sa.ResourceID).First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing service account %s: %w", sa.Email, err)
			}

			diff := DiffServiceAccount(existing, &sa)

			if !diff.HasAnyChange() && existing != nil {
				if err := tx.Model(&bronze.GCPIAMServiceAccount{}).
					Where("resource_id = ?", sa.ResourceID).
					Update("collected_at", sa.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for service account %s: %w", sa.Email, err)
				}
				continue
			}

			if err := tx.Save(&sa).Error; err != nil {
				return fmt.Errorf("failed to upsert service account %s: %w", sa.Email, err)
			}

			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &sa, now); err != nil {
					return fmt.Errorf("failed to create history for service account %s: %w", sa.Email, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &sa, diff, now); err != nil {
					return fmt.Errorf("failed to update history for service account %s: %w", sa.Email, err)
				}
			}
		}
		return nil
	})
}

func (s *Service) DeleteStaleServiceAccounts(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stale []bronze.GCPIAMServiceAccount
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&stale).Error; err != nil {
			return err
		}

		for _, sa := range stale {
			if err := s.history.CloseHistory(tx, sa.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for service account %s: %w", sa.ResourceID, err)
			}
			if err := tx.Delete(&sa).Error; err != nil {
				return fmt.Errorf("failed to delete service account %s: %w", sa.ResourceID, err)
			}
		}
		return nil
	})
}
