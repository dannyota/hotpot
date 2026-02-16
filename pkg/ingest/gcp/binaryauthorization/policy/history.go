package policy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpbinaryauthorizationpolicy"
)

// HistoryService manages Binary Authorization policy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Binary Authorization policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *PolicyData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPBinaryAuthorizationPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetDescription(data.Description).
		SetGlobalPolicyEvaluationMode(data.GlobalPolicyEvaluationMode).
		SetDefaultAdmissionRuleJSON(data.DefaultAdmissionRuleJSON).
		SetClusterAdmissionRulesJSON(data.ClusterAdmissionRulesJSON).
		SetKubeNamespaceAdmissionRulesJSON(data.KubeNamespaceAdmissionRulesJSON).
		SetIstioServiceIdentityAdmissionRulesJSON(data.IstioServiceIdentityAdmissionRulesJSON).
		SetUpdateTime(data.UpdateTime).
		SetEtag(data.Etag).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create binary authorization policy history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Binary Authorization policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPBinaryAuthorizationPolicy, new *PolicyData, diff *PolicyDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBinaryAuthorizationPolicy.Query().
		Where(
			bronzehistorygcpbinaryauthorizationpolicy.ResourceID(old.ID),
			bronzehistorygcpbinaryauthorizationpolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current binary authorization policy history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPBinaryAuthorizationPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current binary authorization policy history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPBinaryAuthorizationPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetDescription(new.Description).
			SetGlobalPolicyEvaluationMode(new.GlobalPolicyEvaluationMode).
			SetDefaultAdmissionRuleJSON(new.DefaultAdmissionRuleJSON).
			SetClusterAdmissionRulesJSON(new.ClusterAdmissionRulesJSON).
			SetKubeNamespaceAdmissionRulesJSON(new.KubeNamespaceAdmissionRulesJSON).
			SetIstioServiceIdentityAdmissionRulesJSON(new.IstioServiceIdentityAdmissionRulesJSON).
			SetUpdateTime(new.UpdateTime).
			SetEtag(new.Etag).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new binary authorization policy history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Binary Authorization policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPBinaryAuthorizationPolicy.Query().
		Where(
			bronzehistorygcpbinaryauthorizationpolicy.ResourceID(resourceID),
			bronzehistorygcpbinaryauthorizationpolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current binary authorization policy history: %w", err)
	}

	err = tx.BronzeHistoryGCPBinaryAuthorizationPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close binary authorization policy history: %w", err)
	}

	return nil
}
