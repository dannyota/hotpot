package securitypolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputesecuritypolicy"
)

// Service handles GCP Compute security policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new security policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for security policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of security policy ingestion.
type IngestResult struct {
	ProjectID           string
	SecurityPolicyCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches security policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch security policies from GCP
	policies, err := s.client.ListSecurityPolicies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list security policies: %w", err)
	}

	// Convert to data structs
	policyDataList := make([]*SecurityPolicyData, 0, len(policies))
	for _, p := range policies {
		data, err := ConvertSecurityPolicy(p, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert security policy: %w", err)
		}
		policyDataList = append(policyDataList, data)
	}

	// Save to database
	if err := s.saveSecurityPolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save security policies: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		SecurityPolicyCount: len(policyDataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSecurityPolicies saves security policies to the database with history tracking.
func (s *Service) saveSecurityPolicies(ctx context.Context, policies []*SecurityPolicyData) error {
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
		// Load existing security policy
		existing, err := tx.BronzeGCPComputeSecurityPolicy.Query().
			Where(bronzegcpcomputesecuritypolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing security policy %s: %w", policyData.Name, err)
		}

		// Compute diff
		diff := DiffSecurityPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeSecurityPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for security policy %s: %w", policyData.Name, err)
			}
			continue
		}

		// Create or update security policy
		if existing == nil {
			// Create new security policy
			create := tx.BronzeGCPComputeSecurityPolicy.Create().
				SetID(policyData.ID).
				SetName(policyData.Name).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				create.SetDescription(policyData.Description)
			}
			if policyData.CreationTimestamp != "" {
				create.SetCreationTimestamp(policyData.CreationTimestamp)
			}
			if policyData.SelfLink != "" {
				create.SetSelfLink(policyData.SelfLink)
			}
			if policyData.Type != "" {
				create.SetType(policyData.Type)
			}
			if policyData.Fingerprint != "" {
				create.SetFingerprint(policyData.Fingerprint)
			}
			if policyData.RulesJSON != nil {
				create.SetRulesJSON(policyData.RulesJSON)
			}
			if policyData.AssociationsJSON != nil {
				create.SetAssociationsJSON(policyData.AssociationsJSON)
			}
			if policyData.AdaptiveProtectionConfigJSON != nil {
				create.SetAdaptiveProtectionConfigJSON(policyData.AdaptiveProtectionConfigJSON)
			}
			if policyData.AdvancedOptionsConfigJSON != nil {
				create.SetAdvancedOptionsConfigJSON(policyData.AdvancedOptionsConfigJSON)
			}
			if policyData.DdosProtectionConfigJSON != nil {
				create.SetDdosProtectionConfigJSON(policyData.DdosProtectionConfigJSON)
			}
			if policyData.RecaptchaOptionsConfigJSON != nil {
				create.SetRecaptchaOptionsConfigJSON(policyData.RecaptchaOptionsConfigJSON)
			}
			if policyData.LabelsJSON != nil {
				create.SetLabelsJSON(policyData.LabelsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create security policy %s: %w", policyData.Name, err)
			}
		} else {
			// Update existing security policy
			update := tx.BronzeGCPComputeSecurityPolicy.UpdateOneID(policyData.ID).
				SetName(policyData.Name).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				update.SetDescription(policyData.Description)
			}
			if policyData.CreationTimestamp != "" {
				update.SetCreationTimestamp(policyData.CreationTimestamp)
			}
			if policyData.SelfLink != "" {
				update.SetSelfLink(policyData.SelfLink)
			}
			if policyData.Type != "" {
				update.SetType(policyData.Type)
			}
			if policyData.Fingerprint != "" {
				update.SetFingerprint(policyData.Fingerprint)
			}
			if policyData.RulesJSON != nil {
				update.SetRulesJSON(policyData.RulesJSON)
			}
			if policyData.AssociationsJSON != nil {
				update.SetAssociationsJSON(policyData.AssociationsJSON)
			}
			if policyData.AdaptiveProtectionConfigJSON != nil {
				update.SetAdaptiveProtectionConfigJSON(policyData.AdaptiveProtectionConfigJSON)
			}
			if policyData.AdvancedOptionsConfigJSON != nil {
				update.SetAdvancedOptionsConfigJSON(policyData.AdvancedOptionsConfigJSON)
			}
			if policyData.DdosProtectionConfigJSON != nil {
				update.SetDdosProtectionConfigJSON(policyData.DdosProtectionConfigJSON)
			}
			if policyData.RecaptchaOptionsConfigJSON != nil {
				update.SetRecaptchaOptionsConfigJSON(policyData.RecaptchaOptionsConfigJSON)
			}
			if policyData.LabelsJSON != nil {
				update.SetLabelsJSON(policyData.LabelsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update security policy %s: %w", policyData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for security policy %s: %w", policyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for security policy %s: %w", policyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSecurityPolicies removes security policies that were not collected in the latest run.
// Also closes history records for deleted security policies.
func (s *Service) DeleteStaleSecurityPolicies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale security policies
	stalePolicies, err := tx.BronzeGCPComputeSecurityPolicy.Query().
		Where(
			bronzegcpcomputesecuritypolicy.ProjectID(projectID),
			bronzegcpcomputesecuritypolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale security policy
	for _, policy := range stalePolicies {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for security policy %s: %w", policy.ID, err)
		}

		// Delete security policy
		if err := tx.BronzeGCPComputeSecurityPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete security policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
