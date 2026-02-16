package accesspolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpaccesscontextmanageraccesspolicy"
)

// Service handles access policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new access policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of access policy ingestion.
type IngestResult struct {
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches access policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch access policies from GCP
	rawPolicies, err := s.client.ListAccessPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list access policies: %w", err)
	}

	// Convert to policy data
	policyDataList := make([]*AccessPolicyData, 0, len(rawPolicies))
	for _, raw := range rawPolicies {
		data := ConvertAccessPolicy(raw.OrgName, raw.AccessPolicy, collectedAt)
		policyDataList = append(policyDataList, data)
	}

	// Save to database
	if err := s.savePolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save access policies: %w", err)
	}

	return &IngestResult{
		PolicyCount:    len(policyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves access policies to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*AccessPolicyData) error {
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

	for _, policyData := range policies {
		// Load existing policy
		existing, err := tx.BronzeGCPAccessContextManagerAccessPolicy.Query().
			Where(bronzegcpaccesscontextmanageraccesspolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing access policy %s: %w", policyData.ID, err)
		}

		// Compute diff
		diff := DiffAccessPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPAccessContextManagerAccessPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for access policy %s: %w", policyData.ID, err)
			}
			continue
		}

		// Create or update policy
		if existing == nil {
			create := tx.BronzeGCPAccessContextManagerAccessPolicy.Create().
				SetID(policyData.ID).
				SetParent(policyData.Parent).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Title != "" {
				create.SetTitle(policyData.Title)
			}
			if policyData.Etag != "" {
				create.SetEtag(policyData.Etag)
			}
			if policyData.ScopesJSON != nil {
				create.SetScopesJSON(policyData.ScopesJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create access policy %s: %w", policyData.ID, err)
			}
		} else {
			update := tx.BronzeGCPAccessContextManagerAccessPolicy.UpdateOneID(policyData.ID).
				SetParent(policyData.Parent).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Title != "" {
				update.SetTitle(policyData.Title)
			}
			if policyData.Etag != "" {
				update.SetEtag(policyData.Etag)
			}
			if policyData.ScopesJSON != nil {
				update.SetScopesJSON(policyData.ScopesJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update access policy %s: %w", policyData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for access policy %s: %w", policyData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for access policy %s: %w", policyData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes access policies that were not collected in the latest run.
func (s *Service) DeleteStalePolicies(ctx context.Context, collectedAt time.Time) error {
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

	stalePolicies, err := tx.BronzeGCPAccessContextManagerAccessPolicy.Query().
		Where(bronzegcpaccesscontextmanageraccesspolicy.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, policy := range stalePolicies {
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for access policy %s: %w", policy.ID, err)
		}

		if err := tx.BronzeGCPAccessContextManagerAccessPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete access policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
