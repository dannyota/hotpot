package policy

import (
	"context"
	"fmt"
	"time"

	entorgpolicy "github.com/dannyota/hotpot/pkg/storage/ent/gcp/orgpolicy"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/orgpolicy/bronzehistorygcporgpolicypolicy"
)

type HistoryService struct {
	entClient *entorgpolicy.Client
}

func NewHistoryService(entClient *entorgpolicy.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *entorgpolicy.Tx, data *PolicyData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPOrgPolicyPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetEtag(data.Etag).
		SetSpec(data.Spec).
		SetDryRunSpec(data.DryRunSpec).
		SetOrganizationID(data.OrganizationID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create org policy history: %w", err)
	}

	return nil
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entorgpolicy.Tx, old *entorgpolicy.BronzeGCPOrgPolicyPolicy, new *PolicyData, diff *PolicyDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyPolicy.Query().
		Where(
			bronzehistorygcporgpolicypolicy.ResourceID(old.ID),
			bronzehistorygcporgpolicypolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current org policy history: %w", err)
	}

	if diff.IsChanged {
		err = tx.BronzeHistoryGCPOrgPolicyPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current org policy history: %w", err)
		}

		_, err := tx.BronzeHistoryGCPOrgPolicyPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetEtag(new.Etag).
			SetSpec(new.Spec).
			SetDryRunSpec(new.DryRunSpec).
			SetOrganizationID(new.OrganizationID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new org policy history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *entorgpolicy.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPOrgPolicyPolicy.Query().
		Where(
			bronzehistorygcporgpolicypolicy.ResourceID(resourceID),
			bronzehistorygcporgpolicypolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entorgpolicy.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current org policy history: %w", err)
	}

	err = tx.BronzeHistoryGCPOrgPolicyPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close org policy history: %w", err)
	}

	return nil
}
