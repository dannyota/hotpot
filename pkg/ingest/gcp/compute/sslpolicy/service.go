package sslpolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputesslpolicy"
)

// Service handles GCP Compute SSL policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new SSL policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for SSL policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of SSL policy ingestion.
type IngestResult struct {
	ProjectID      string
	SslPolicyCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches SSL policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch SSL policies from GCP
	sslPolicies, err := s.client.ListSslPolicies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list SSL policies: %w", err)
	}

	// Convert to data structs
	sslPolicyDataList := make([]*SslPolicyData, 0, len(sslPolicies))
	for _, p := range sslPolicies {
		data, err := ConvertSslPolicy(p, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert SSL policy: %w", err)
		}
		sslPolicyDataList = append(sslPolicyDataList, data)
	}

	// Save to database
	if err := s.saveSslPolicies(ctx, sslPolicyDataList); err != nil {
		return nil, fmt.Errorf("failed to save SSL policies: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		SslPolicyCount: len(sslPolicyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSslPolicies saves SSL policies to the database with history tracking.
func (s *Service) saveSslPolicies(ctx context.Context, policies []*SslPolicyData) error {
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
		// Load existing SSL policy
		existing, err := tx.BronzeGCPComputeSslPolicy.Query().
			Where(bronzegcpcomputesslpolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing SSL policy %s: %w", policyData.Name, err)
		}

		// Compute diff
		diff := DiffSslPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeSslPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for SSL policy %s: %w", policyData.Name, err)
			}
			continue
		}

		// Create or update SSL policy
		if existing == nil {
			// Create new SSL policy
			create := tx.BronzeGCPComputeSslPolicy.Create().
				SetID(policyData.ID).
				SetName(policyData.Name).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				create.SetDescription(policyData.Description)
			}
			if policyData.SelfLink != "" {
				create.SetSelfLink(policyData.SelfLink)
			}
			if policyData.CreationTimestamp != "" {
				create.SetCreationTimestamp(policyData.CreationTimestamp)
			}
			if policyData.Profile != "" {
				create.SetProfile(policyData.Profile)
			}
			if policyData.MinTlsVersion != "" {
				create.SetMinTLSVersion(policyData.MinTlsVersion)
			}
			if policyData.Fingerprint != "" {
				create.SetFingerprint(policyData.Fingerprint)
			}
			if policyData.CustomFeaturesJSON != nil {
				create.SetCustomFeaturesJSON(policyData.CustomFeaturesJSON)
			}
			if policyData.EnabledFeaturesJSON != nil {
				create.SetEnabledFeaturesJSON(policyData.EnabledFeaturesJSON)
			}
			if policyData.WarningsJSON != nil {
				create.SetWarningsJSON(policyData.WarningsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create SSL policy %s: %w", policyData.Name, err)
			}
		} else {
			// Update existing SSL policy
			update := tx.BronzeGCPComputeSslPolicy.UpdateOneID(policyData.ID).
				SetName(policyData.Name).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				update.SetDescription(policyData.Description)
			}
			if policyData.SelfLink != "" {
				update.SetSelfLink(policyData.SelfLink)
			}
			if policyData.CreationTimestamp != "" {
				update.SetCreationTimestamp(policyData.CreationTimestamp)
			}
			if policyData.Profile != "" {
				update.SetProfile(policyData.Profile)
			}
			if policyData.MinTlsVersion != "" {
				update.SetMinTLSVersion(policyData.MinTlsVersion)
			}
			if policyData.Fingerprint != "" {
				update.SetFingerprint(policyData.Fingerprint)
			}
			if policyData.CustomFeaturesJSON != nil {
				update.SetCustomFeaturesJSON(policyData.CustomFeaturesJSON)
			}
			if policyData.EnabledFeaturesJSON != nil {
				update.SetEnabledFeaturesJSON(policyData.EnabledFeaturesJSON)
			}
			if policyData.WarningsJSON != nil {
				update.SetWarningsJSON(policyData.WarningsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update SSL policy %s: %w", policyData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for SSL policy %s: %w", policyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for SSL policy %s: %w", policyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSslPolicies removes SSL policies that were not collected in the latest run.
// Also closes history records for deleted SSL policies.
func (s *Service) DeleteStaleSslPolicies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale SSL policies
	stalePolicies, err := tx.BronzeGCPComputeSslPolicy.Query().
		Where(
			bronzegcpcomputesslpolicy.ProjectID(projectID),
			bronzegcpcomputesslpolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale SSL policy
	for _, policy := range stalePolicies {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for SSL policy %s: %w", policy.ID, err)
		}

		// Delete SSL policy
		if err := tx.BronzeGCPComputeSslPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete SSL policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
