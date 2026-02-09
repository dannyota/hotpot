package account

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1account"
)

// Service handles SentinelOne account ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new account ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of account ingestion.
type IngestResult struct {
	AccountCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches accounts from SentinelOne and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	accounts, err := s.client.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	dataList := make([]*AccountData, 0, len(accounts))
	for _, acct := range accounts {
		dataList = append(dataList, ConvertAccount(acct, collectedAt))
	}

	if err := s.saveAccounts(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save accounts: %w", err)
	}

	return &IngestResult{
		AccountCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveAccounts(ctx context.Context, accounts []*AccountData) error {
	if len(accounts) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range accounts {
		existing, err := tx.BronzeS1Account.Query().
			Where(bronzes1account.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing account %s: %w", data.ResourceID, err)
		}

		diff := DiffAccountData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1Account.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for account %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1Account.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetState(data.State).
				SetAccountType(data.AccountType).
				SetUnlimitedExpiration(data.UnlimitedExpiration).
				SetActiveAgents(data.ActiveAgents).
				SetTotalLicenses(data.TotalLicenses).
				SetUsageType(data.UsageType).
				SetBillingMode(data.BillingMode).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetNumberOfSites(data.NumberOfSites).
				SetExternalID(data.ExternalID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.APIUpdatedAt != nil {
				create.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}
			if data.Expiration != nil {
				create.SetExpiration(*data.Expiration)
			}
			if data.LicensesJSON != nil {
				create.SetLicensesJSON(data.LicensesJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create account %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for account %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1Account.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetState(data.State).
				SetAccountType(data.AccountType).
				SetUnlimitedExpiration(data.UnlimitedExpiration).
				SetActiveAgents(data.ActiveAgents).
				SetTotalLicenses(data.TotalLicenses).
				SetUsageType(data.UsageType).
				SetBillingMode(data.BillingMode).
				SetCreator(data.Creator).
				SetCreatorID(data.CreatorID).
				SetNumberOfSites(data.NumberOfSites).
				SetExternalID(data.ExternalID).
				SetCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.APIUpdatedAt != nil {
				update.SetAPIUpdatedAt(*data.APIUpdatedAt)
			}
			if data.Expiration != nil {
				update.SetExpiration(*data.Expiration)
			}
			if data.LicensesJSON != nil {
				update.SetLicensesJSON(data.LicensesJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update account %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for account %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes accounts that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeS1Account.Query().
		Where(bronzes1account.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, acct := range stale {
		if err := s.history.CloseHistory(ctx, tx, acct.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for account %s: %w", acct.ID, err)
		}

		if err := tx.BronzeS1Account.DeleteOne(acct).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete account %s: %w", acct.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
