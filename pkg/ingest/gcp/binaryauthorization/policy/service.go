package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpbinaryauthorizationpolicy"
)

// Service handles Binary Authorization policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Binary Authorization policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of Binary Authorization policy ingestion.
type IngestResult struct {
	ProjectID      string
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches the Binary Authorization policy from GCP and stores it in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawPolicy, err := s.client.GetPolicy(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	policyData := ConvertPolicy(rawPolicy, params.ProjectID, collectedAt)

	if err := s.savePolicy(ctx, policyData); err != nil {
		return nil, fmt.Errorf("failed to save policy: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		PolicyCount:    1,
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicy saves a Binary Authorization policy to the database with history tracking.
func (s *Service) savePolicy(ctx context.Context, policyData *PolicyData) error {
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

	// Load existing policy
	existing, err := tx.BronzeGCPBinaryAuthorizationPolicy.Query().
		Where(bronzegcpbinaryauthorizationpolicy.ID(policyData.ID)).
		First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		tx.Rollback()
		return fmt.Errorf("failed to load existing policy %s: %w", policyData.ID, err)
	}

	// Compute diff
	diff := DiffPolicyData(existing, policyData)

	// Skip if no changes
	if !diff.HasAnyChange() && existing != nil {
		if err := tx.BronzeGCPBinaryAuthorizationPolicy.UpdateOneID(policyData.ID).
			SetCollectedAt(policyData.CollectedAt).
			Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update collected_at for policy %s: %w", policyData.ID, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}

	// Create or update policy
	if existing == nil {
		create := tx.BronzeGCPBinaryAuthorizationPolicy.Create().
			SetID(policyData.ID).
			SetProjectID(policyData.ProjectID).
			SetCollectedAt(policyData.CollectedAt).
			SetFirstCollectedAt(policyData.CollectedAt).
			SetGlobalPolicyEvaluationMode(policyData.GlobalPolicyEvaluationMode)

		if policyData.Description != "" {
			create.SetDescription(policyData.Description)
		}
		if policyData.DefaultAdmissionRuleJSON != nil {
			create.SetDefaultAdmissionRuleJSON(policyData.DefaultAdmissionRuleJSON)
		}
		if policyData.ClusterAdmissionRulesJSON != nil {
			create.SetClusterAdmissionRulesJSON(policyData.ClusterAdmissionRulesJSON)
		}
		if policyData.KubeNamespaceAdmissionRulesJSON != nil {
			create.SetKubeNamespaceAdmissionRulesJSON(policyData.KubeNamespaceAdmissionRulesJSON)
		}
		if policyData.IstioServiceIdentityAdmissionRulesJSON != nil {
			create.SetIstioServiceIdentityAdmissionRulesJSON(policyData.IstioServiceIdentityAdmissionRulesJSON)
		}
		if policyData.UpdateTime != "" {
			create.SetUpdateTime(policyData.UpdateTime)
		}
		if policyData.Etag != "" {
			create.SetEtag(policyData.Etag)
		}

		_, err = create.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create policy %s: %w", policyData.ID, err)
		}
	} else {
		update := tx.BronzeGCPBinaryAuthorizationPolicy.UpdateOneID(policyData.ID).
			SetProjectID(policyData.ProjectID).
			SetCollectedAt(policyData.CollectedAt).
			SetGlobalPolicyEvaluationMode(policyData.GlobalPolicyEvaluationMode)

		if policyData.Description != "" {
			update.SetDescription(policyData.Description)
		}
		if policyData.DefaultAdmissionRuleJSON != nil {
			update.SetDefaultAdmissionRuleJSON(policyData.DefaultAdmissionRuleJSON)
		}
		if policyData.ClusterAdmissionRulesJSON != nil {
			update.SetClusterAdmissionRulesJSON(policyData.ClusterAdmissionRulesJSON)
		}
		if policyData.KubeNamespaceAdmissionRulesJSON != nil {
			update.SetKubeNamespaceAdmissionRulesJSON(policyData.KubeNamespaceAdmissionRulesJSON)
		}
		if policyData.IstioServiceIdentityAdmissionRulesJSON != nil {
			update.SetIstioServiceIdentityAdmissionRulesJSON(policyData.IstioServiceIdentityAdmissionRulesJSON)
		}
		if policyData.UpdateTime != "" {
			update.SetUpdateTime(policyData.UpdateTime)
		}
		if policyData.Etag != "" {
			update.SetEtag(policyData.Etag)
		}

		_, err = update.Save(ctx)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update policy %s: %w", policyData.ID, err)
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

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes policies that were not collected in the latest run.
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

	stalePolicies, err := tx.BronzeGCPBinaryAuthorizationPolicy.Query().
		Where(
			bronzegcpbinaryauthorizationpolicy.ProjectID(projectID),
			bronzegcpbinaryauthorizationpolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, pol := range stalePolicies {
		if err := s.history.CloseHistory(ctx, tx, pol.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for policy %s: %w", pol.ID, err)
		}

		if err := tx.BronzeGCPBinaryAuthorizationPolicy.DeleteOne(pol).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete policy %s: %w", pol.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
