package orgiampolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporgiampolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporgiampolicybinding"
)

// Service handles GCP organization IAM policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new organization IAM policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of organization IAM policy ingestion.
type IngestResult struct {
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches organization IAM policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch organization IAM policies from GCP
	rawPolicies, err := s.client.ListOrgIamPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list org IAM policies: %w", err)
	}

	// Convert to policy data
	policyDataList := make([]*OrgIamPolicyData, 0, len(rawPolicies))
	for _, raw := range rawPolicies {
		data, err := ConvertOrgIamPolicy(raw.OrgName, raw.Policy, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert org IAM policy: %w", err)
		}
		policyDataList = append(policyDataList, data)
	}

	// Save to database
	if err := s.savePolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save org IAM policies: %w", err)
	}

	return &IngestResult{
		PolicyCount:    len(policyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves organization IAM policies to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*OrgIamPolicyData) error {
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
		// Load existing policy with bindings
		existing, err := tx.BronzeGCPOrgIamPolicy.Query().
			Where(bronzegcporgiampolicy.ID(policyData.ID)).
			WithBindings().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing policy %s: %w", policyData.ID, err)
		}

		// Compute diff
		diff := DiffOrgIamPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPOrgIamPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for policy %s: %w", policyData.ID, err)
			}
			continue
		}

		// Delete old bindings if updating
		if existing != nil {
			_, err := tx.BronzeGCPOrgIamPolicyBinding.Delete().
				Where(bronzegcporgiampolicybinding.HasPolicyWith(bronzegcporgiampolicy.ID(policyData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old bindings for policy %s: %w", policyData.ID, err)
			}
		}

		// Create or update policy
		var savedPolicy *ent.BronzeGCPOrgIamPolicy
		if existing == nil {
			// Create new policy
			create := tx.BronzeGCPOrgIamPolicy.Create().
				SetID(policyData.ID).
				SetResourceName(policyData.ResourceName).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Etag != "" {
				create.SetEtag(policyData.Etag)
			}
			if policyData.Version != 0 {
				create.SetVersion(policyData.Version)
			}

			savedPolicy, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create policy %s: %w", policyData.ID, err)
			}
		} else {
			// Update existing policy
			update := tx.BronzeGCPOrgIamPolicy.UpdateOneID(policyData.ID).
				SetResourceName(policyData.ResourceName).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Etag != "" {
				update.SetEtag(policyData.Etag)
			}
			if policyData.Version != 0 {
				update.SetVersion(policyData.Version)
			}

			savedPolicy, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update policy %s: %w", policyData.ID, err)
			}
		}

		// Create bindings
		for _, binding := range policyData.Bindings {
			create := tx.BronzeGCPOrgIamPolicyBinding.Create().
				SetRole(binding.Role).
				SetPolicy(savedPolicy)

			if binding.MembersJSON != nil {
				create.SetMembersJSON(binding.MembersJSON)
			}
			if binding.ConditionJSON != nil {
				create.SetConditionJSON(binding.ConditionJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create binding for policy %s: %w", policyData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for policy %s: %w", policyData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for policy %s: %w", policyData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes policies that were not collected in the latest run.
// Also closes history records for deleted policies.
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

	// Find stale policies
	stalePolicies, err := tx.BronzeGCPOrgIamPolicy.Query().
		Where(
			bronzegcporgiampolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale policy
	for _, policy := range stalePolicies {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for policy %s: %w", policy.ID, err)
		}

		// Delete policy (bindings will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPOrgIamPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
