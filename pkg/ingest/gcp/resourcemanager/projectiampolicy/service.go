package projectiampolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpprojectiampolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpprojectiampolicybinding"
)

// Service handles GCP project IAM policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new project IAM policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of project IAM policy ingestion.
type IngestResult struct {
	ProjectID      string
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches the project IAM policy from GCP and stores it in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch project IAM policy from GCP
	raw, err := s.client.GetProjectIamPolicy(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project IAM policy: %w", err)
	}

	// Convert to policy data
	policyData, err := ConvertProjectIamPolicy(raw, collectedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert project IAM policy: %w", err)
	}

	// Save to database
	if err := s.savePolicies(ctx, []*ProjectIamPolicyData{policyData}); err != nil {
		return nil, fmt.Errorf("failed to save project IAM policy: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		PolicyCount:    1,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves project IAM policies to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*ProjectIamPolicyData) error {
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
		existing, err := tx.BronzeGCPProjectIamPolicy.Query().
			Where(bronzegcpprojectiampolicy.ID(policyData.ID)).
			WithBindings().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing policy %s: %w", policyData.ID, err)
		}

		// Compute diff
		diff := DiffProjectIamPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPProjectIamPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for policy %s: %w", policyData.ID, err)
			}
			continue
		}

		// Delete old bindings if updating
		if existing != nil {
			_, err := tx.BronzeGCPProjectIamPolicyBinding.Delete().
				Where(bronzegcpprojectiampolicybinding.HasPolicyWith(bronzegcpprojectiampolicy.ID(policyData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old bindings for policy %s: %w", policyData.ID, err)
			}
		}

		// Create or update policy
		var savedPolicy *ent.BronzeGCPProjectIamPolicy
		if existing == nil {
			// Create new policy
			create := tx.BronzeGCPProjectIamPolicy.Create().
				SetID(policyData.ID).
				SetResourceName(policyData.ResourceName).
				SetProjectID(policyData.ProjectID).
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
			update := tx.BronzeGCPProjectIamPolicy.UpdateOneID(policyData.ID).
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
			create := tx.BronzeGCPProjectIamPolicyBinding.Create().
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

	// Find stale policies
	stalePolicies, err := tx.BronzeGCPProjectIamPolicy.Query().
		Where(
			bronzegcpprojectiampolicy.ProjectID(projectID),
			bronzegcpprojectiampolicy.CollectedAtLT(collectedAt),
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
		if err := tx.BronzeGCPProjectIamPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
