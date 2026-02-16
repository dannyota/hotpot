package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcporgpolicypolicy"
)

type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

type IngestResult struct {
	PolicyCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

func (s *Service) Ingest(ctx context.Context) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawPolicies, err := s.client.ListPolicies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list policies: %w", err)
	}

	policyDataList := make([]*PolicyData, 0, len(rawPolicies))
	for _, raw := range rawPolicies {
		data := ConvertPolicy(raw.OrgName, raw.Policy, collectedAt)
		policyDataList = append(policyDataList, data)
	}

	if err := s.savePolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save policies: %w", err)
	}

	return &IngestResult{
		PolicyCount:    len(policyDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

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
		existing, err := tx.BronzeGCPOrgPolicyPolicy.Query().
			Where(bronzegcporgpolicypolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing policy %s: %w", policyData.ID, err)
		}

		diff := DiffPolicyData(existing, policyData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPOrgPolicyPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for policy %s: %w", policyData.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPOrgPolicyPolicy.Create().
				SetID(policyData.ID).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.Etag != "" {
				create.SetEtag(policyData.Etag)
			}
			if policyData.Spec != nil {
				create.SetSpec(policyData.Spec)
			}
			if policyData.DryRunSpec != nil {
				create.SetDryRunSpec(policyData.DryRunSpec)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create policy %s: %w", policyData.ID, err)
			}
		} else {
			update := tx.BronzeGCPOrgPolicyPolicy.UpdateOneID(policyData.ID).
				SetOrganizationID(policyData.OrganizationID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.Etag != "" {
				update.SetEtag(policyData.Etag)
			}
			if policyData.Spec != nil {
				update.SetSpec(policyData.Spec)
			}
			if policyData.DryRunSpec != nil {
				update.SetDryRunSpec(policyData.DryRunSpec)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update policy %s: %w", policyData.ID, err)
			}
		}

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

	stalePolicies, err := tx.BronzeGCPOrgPolicyPolicy.Query().
		Where(bronzegcporgpolicypolicy.CollectedAtLT(collectedAt)).
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

		if err := tx.BronzeGCPOrgPolicyPolicy.DeleteOne(pol).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete policy %s: %w", pol.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
