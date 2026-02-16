package iampolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpiapiampolicy"
)

// Service handles IAP IAM policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new IAP IAM policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for IAP IAM policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of IAP IAM policy ingestion.
type IngestResult struct {
	ProjectID      string
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches IAP IAM policy from GCP and stores it in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch IAP IAM policy from GCP (single object per project)
	raw, err := s.client.GetIAMPolicy(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get IAP IAM policy: %w", err)
	}

	// Convert to data struct
	data, err := ConvertIAMPolicy(raw, params.ProjectID, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert IAP IAM policy: %w", err)
	}

	// Save to database
	dataList := []*IAMPolicyData{data}
	if err := s.savePolicies(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save IAP IAM policy: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		PolicyCount:    1,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves IAP IAM policies to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*IAMPolicyData) error {
	if len(policies) == 0 {
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

	for _, data := range policies {
		// Load existing policy
		existing, err := tx.BronzeGCPIAPIAMPolicy.Query().
			Where(bronzegcpiapiampolicy.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing IAP IAM policy %s: %w", data.ID, err)
		}

		// Compute diff
		diff := DiffIAMPolicyData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPIAPIAMPolicy.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for IAP IAM policy %s: %w", data.ID, err)
			}
			continue
		}

		// Create or update policy
		if existing == nil {
			create := tx.BronzeGCPIAPIAMPolicy.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.Etag != "" {
				create.SetEtag(data.Etag)
			}
			if data.Version != 0 {
				create.SetVersion(data.Version)
			}
			if data.BindingsJSON != nil {
				create.SetBindingsJSON(data.BindingsJSON)
			}
			if data.AuditConfigsJSON != nil {
				create.SetAuditConfigsJSON(data.AuditConfigsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create IAP IAM policy %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGCPIAPIAMPolicy.UpdateOneID(data.ID).
				SetName(data.Name).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.Etag != "" {
				update.SetEtag(data.Etag)
			}
			if data.Version != 0 {
				update.SetVersion(data.Version)
			}
			if data.BindingsJSON != nil {
				update.SetBindingsJSON(data.BindingsJSON)
			}
			if data.AuditConfigsJSON != nil {
				update.SetAuditConfigsJSON(data.AuditConfigsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update IAP IAM policy %s: %w", data.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for IAP IAM policy %s: %w", data.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for IAP IAM policy %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes IAP IAM policies that were not collected in the latest run.
func (s *Service) DeleteStalePolicies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stalePolicies, err := tx.BronzeGCPIAPIAMPolicy.Query().
		Where(
			bronzegcpiapiampolicy.ProjectID(projectID),
			bronzegcpiapiampolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, stale := range stalePolicies {
		if err := s.history.CloseHistory(ctx, tx, stale.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for IAP IAM policy %s: %w", stale.ID, err)
		}

		if err := tx.BronzeGCPIAPIAMPolicy.DeleteOne(stale).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete IAP IAM policy %s: %w", stale.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
