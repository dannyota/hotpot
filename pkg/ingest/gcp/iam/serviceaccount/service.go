package serviceaccount

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpiamserviceaccount"
)

type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
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

	accountDataList := make([]*ServiceAccountData, 0, len(accounts))
	for _, sa := range accounts {
		accountDataList = append(accountDataList, ConvertServiceAccount(sa, params.ProjectID, collectedAt))
	}

	if err := s.saveServiceAccounts(ctx, accountDataList); err != nil {
		return nil, fmt.Errorf("failed to save service accounts: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		ServiceAccountCount: len(accountDataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveServiceAccounts(ctx context.Context, accounts []*ServiceAccountData) error {
	if len(accounts) == 0 {
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

	for _, saData := range accounts {
		existing, err := tx.BronzeGCPIAMServiceAccount.Query().
			Where(bronzegcpiamserviceaccount.ID(saData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing service account %s: %w", saData.Email, err)
		}

		diff := DiffServiceAccountData(existing, saData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPIAMServiceAccount.UpdateOneID(saData.ResourceID).
				SetCollectedAt(saData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for service account %s: %w", saData.Email, err)
			}
			continue
		}

		if existing == nil {
			// Create new service account
			_, err := tx.BronzeGCPIAMServiceAccount.Create().
				SetID(saData.ResourceID).
				SetName(saData.Name).
				SetEmail(saData.Email).
				SetDisplayName(saData.DisplayName).
				SetDescription(saData.Description).
				SetOauth2ClientID(saData.Oauth2ClientId).
				SetDisabled(saData.Disabled).
				SetEtag(saData.Etag).
				SetProjectID(saData.ProjectID).
				SetCollectedAt(saData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create service account %s: %w", saData.Email, err)
			}
		} else {
			// Update existing service account
			_, err := tx.BronzeGCPIAMServiceAccount.UpdateOneID(saData.ResourceID).
				SetName(saData.Name).
				SetEmail(saData.Email).
				SetDisplayName(saData.DisplayName).
				SetDescription(saData.Description).
				SetOauth2ClientID(saData.Oauth2ClientId).
				SetDisabled(saData.Disabled).
				SetEtag(saData.Etag).
				SetProjectID(saData.ProjectID).
				SetCollectedAt(saData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update service account %s: %w", saData.Email, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, saData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for service account %s: %w", saData.Email, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, saData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for service account %s: %w", saData.Email, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteStaleServiceAccounts(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPIAMServiceAccount.Query().
		Where(
			bronzegcpiamserviceaccount.ProjectID(projectID),
			bronzegcpiamserviceaccount.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, sa := range stale {
		if err := s.history.CloseHistory(ctx, tx, sa.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for service account %s: %w", sa.ID, err)
		}
		if err := tx.BronzeGCPIAMServiceAccount.DeleteOne(sa).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete service account %s: %w", sa.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
