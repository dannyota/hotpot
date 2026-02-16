package account

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedoaccount"
)

// Service handles DigitalOcean Account ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Account ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Account ingestion.
type IngestResult struct {
	AccountCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches the Account from DigitalOcean and saves it.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiAccount, err := s.client.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	allAccounts := []*AccountData{ConvertAccount(apiAccount, collectedAt)}

	if err := s.saveAccounts(ctx, allAccounts); err != nil {
		return nil, fmt.Errorf("save accounts: %w", err)
	}

	return &IngestResult{
		AccountCount:   len(allAccounts),
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
		existing, err := tx.BronzeDOAccount.Query().
			Where(bronzedoaccount.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing Account %s: %w", data.ResourceID, err)
		}

		diff := DiffAccountData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOAccount.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for Account %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOAccount.Create().
				SetID(data.ResourceID).
				SetEmail(data.Email).
				SetName(data.Name).
				SetStatus(data.Status).
				SetStatusMessage(data.StatusMessage).
				SetDropletLimit(data.DropletLimit).
				SetFloatingIPLimit(data.FloatingIPLimit).
				SetReservedIPLimit(data.ReservedIPLimit).
				SetVolumeLimit(data.VolumeLimit).
				SetEmailVerified(data.EmailVerified).
				SetTeamName(data.TeamName).
				SetTeamUUID(data.TeamUUID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create Account %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for Account %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOAccount.UpdateOneID(data.ResourceID).
				SetEmail(data.Email).
				SetName(data.Name).
				SetStatus(data.Status).
				SetStatusMessage(data.StatusMessage).
				SetDropletLimit(data.DropletLimit).
				SetFloatingIPLimit(data.FloatingIPLimit).
				SetReservedIPLimit(data.ReservedIPLimit).
				SetVolumeLimit(data.VolumeLimit).
				SetEmailVerified(data.EmailVerified).
				SetTeamName(data.TeamName).
				SetTeamUUID(data.TeamUUID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update Account %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for Account %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes Accounts that were not collected in the latest run.
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

	stale, err := tx.BronzeDOAccount.Query().
		Where(bronzedoaccount.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doAccount := range stale {
		if err := s.history.CloseHistory(ctx, tx, doAccount.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for Account %s: %w", doAccount.ID, err)
		}

		if err := tx.BronzeDOAccount.DeleteOne(doAccount).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete Account %s: %w", doAccount.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
