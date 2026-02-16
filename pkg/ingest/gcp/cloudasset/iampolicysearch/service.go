package iampolicysearch

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcloudassetiampolicysearch"
)

// Service handles IAM policy search ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new IAM policy search ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of IAM policy search ingestion.
type IngestResult struct {
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches IAM policy search results from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch IAM policy search results from GCP
	rawPolicies, err := s.client.SearchAllIamPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to search IAM policies: %w", err)
	}

	// Convert to policy data
	policyDataList := make([]*IAMPolicySearchData, 0, len(rawPolicies))
	for _, raw := range rawPolicies {
		data := ConvertIAMPolicySearch(raw.OrgName, raw.Policy, collectedAt)
		policyDataList = append(policyDataList, data)
	}

	// Save to database
	if err := s.savePolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save IAM policy search results: %w", err)
	}

	return &IngestResult{
		PolicyCount:    len(policyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves IAM policy search results to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*IAMPolicySearchData) error {
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
		existing, err := tx.BronzeGCPCloudAssetIAMPolicySearch.Query().
			Where(bronzegcpcloudassetiampolicysearch.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing IAM policy search %s: %w", policyData.ID, err)
		}

		// Compute diff
		diff := DiffIAMPolicySearchData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPCloudAssetIAMPolicySearch.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for IAM policy search %s: %w", policyData.ID, err)
			}
			continue
		}

		// Create or update policy
		if existing == nil {
			create := tx.BronzeGCPCloudAssetIAMPolicySearch.Create().
				SetID(policyData.ID).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.AssetType != "" {
				create.SetAssetType(policyData.AssetType)
			}
			if policyData.Project != "" {
				create.SetProject(policyData.Project)
			}
			if policyData.Organization != "" {
				create.SetOrganization(policyData.Organization)
			}
			if policyData.FoldersJSON != nil {
				create.SetFoldersJSON(policyData.FoldersJSON)
			}
			if policyData.PolicyJSON != nil {
				create.SetPolicyJSON(policyData.PolicyJSON)
			}
			if policyData.ExplanationJSON != nil {
				create.SetExplanationJSON(policyData.ExplanationJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create IAM policy search %s: %w", policyData.ID, err)
			}
		} else {
			update := tx.BronzeGCPCloudAssetIAMPolicySearch.UpdateOneID(policyData.ID).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.AssetType != "" {
				update.SetAssetType(policyData.AssetType)
			}
			if policyData.Project != "" {
				update.SetProject(policyData.Project)
			}
			if policyData.Organization != "" {
				update.SetOrganization(policyData.Organization)
			}
			if policyData.FoldersJSON != nil {
				update.SetFoldersJSON(policyData.FoldersJSON)
			}
			if policyData.PolicyJSON != nil {
				update.SetPolicyJSON(policyData.PolicyJSON)
			}
			if policyData.ExplanationJSON != nil {
				update.SetExplanationJSON(policyData.ExplanationJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update IAM policy search %s: %w", policyData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for IAM policy search %s: %w", policyData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for IAM policy search %s: %w", policyData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes IAM policy search results that were not collected in the latest run.
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

	stalePolicies, err := tx.BronzeGCPCloudAssetIAMPolicySearch.Query().
		Where(bronzegcpcloudassetiampolicysearch.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, policy := range stalePolicies {
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for IAM policy search %s: %w", policy.ID, err)
		}

		if err := tx.BronzeGCPCloudAssetIAMPolicySearch.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete IAM policy search %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
