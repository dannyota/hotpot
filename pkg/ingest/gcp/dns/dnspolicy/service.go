package dnspolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpdnspolicy"
)

// Service handles GCP DNS policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new DNS policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for DNS policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of DNS policy ingestion.
type IngestResult struct {
	ProjectID      string
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches DNS policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	policies, err := s.client.ListPolicies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	policyDataList := make([]*PolicyData, 0, len(policies))
	for _, policy := range policies {
		data, err := ConvertPolicy(policy, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert policy: %w", err)
		}
		policyDataList = append(policyDataList, data)
	}

	if err := s.savePolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save policies: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		PolicyCount:    len(policyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// savePolicies saves DNS policies to the database with history tracking.
func (s *Service) savePolicies(ctx context.Context, policies []*PolicyData) error {
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
		existing, err := tx.BronzeGCPDNSPolicy.Query().
			Where(bronzegcpdnspolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing policy %s: %w", policyData.Name, err)
		}

		diff := DiffPolicyData(existing, policyData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPDNSPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for policy %s: %w", policyData.Name, err)
			}
			continue
		}

		if existing == nil {
			// Create new policy
			create := tx.BronzeGCPDNSPolicy.Create().
				SetID(policyData.ID).
				SetName(policyData.Name).
				SetEnableInboundForwarding(policyData.EnableInboundForwarding).
				SetEnableLogging(policyData.EnableLogging).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				create.SetDescription(policyData.Description)
			}
			if policyData.NetworksJSON != nil {
				create.SetNetworksJSON(policyData.NetworksJSON)
			}
			if policyData.AlternativeNameServerConfigJSON != nil {
				create.SetAlternativeNameServerConfigJSON(policyData.AlternativeNameServerConfigJSON)
			}

			_, err := create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create policy %s: %w", policyData.Name, err)
			}
		} else {
			// Update existing policy
			update := tx.BronzeGCPDNSPolicy.UpdateOneID(policyData.ID).
				SetName(policyData.Name).
				SetEnableInboundForwarding(policyData.EnableInboundForwarding).
				SetEnableLogging(policyData.EnableLogging).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Description != "" {
				update.SetDescription(policyData.Description)
			}
			if policyData.NetworksJSON != nil {
				update.SetNetworksJSON(policyData.NetworksJSON)
			}
			if policyData.AlternativeNameServerConfigJSON != nil {
				update.SetAlternativeNameServerConfigJSON(policyData.AlternativeNameServerConfigJSON)
			}

			_, err := update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update policy %s: %w", policyData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for policy %s: %w", policyData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for policy %s: %w", policyData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStalePolicies removes DNS policies that were not collected in the latest run.
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

	stale, err := tx.BronzeGCPDNSPolicy.Query().
		Where(
			bronzegcpdnspolicy.ProjectID(projectID),
			bronzegcpdnspolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, stalePolicy := range stale {
		if err := s.history.CloseHistory(ctx, tx, stalePolicy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for policy %s: %w", stalePolicy.ID, err)
		}
		if err := tx.BronzeGCPDNSPolicy.DeleteOne(stalePolicy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete policy %s: %w", stalePolicy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
