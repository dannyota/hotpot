package iampolicy

import (
	"context"
	"fmt"
	"time"

	entiap "danny.vn/hotpot/pkg/storage/ent/gcp/iap"
	"danny.vn/hotpot/pkg/storage/ent/gcp/iap/bronzehistorygcpiapiampolicy"
)

// HistoryService manages IAP IAM policy history tracking.
type HistoryService struct {
	entClient *entiap.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entiap.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new IAP IAM policy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entiap.Tx, data *IAMPolicyData, now time.Time) error {
	create := tx.BronzeHistoryGCPIAPIAMPolicy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.Etag != "" {
		create.SetEtag(data.Etag)
	}
	if data.Version != 0 {
		create.SetVersion(data.Version)
	}
	if data.BindingsJSON != nil {
		create.SetBindingsJSON(data.BindingsJSON)
	}
	if data.AuditConfigsJSON != nil {
		create.SetAuditConfigsJSON(data.AuditConfigsJSON)
	}

	_, err := create.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create IAP IAM policy history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed IAP IAM policy.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entiap.Tx, old *entiap.BronzeGCPIAPIAMPolicy, new *IAMPolicyData, diff *IAMPolicyDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPIAPIAMPolicy.Query().
		Where(
			bronzehistorygcpiapiampolicy.ResourceID(old.ID),
			bronzehistorygcpiapiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current IAP IAM policy history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPIAPIAMPolicy.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current IAP IAM policy history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPIAPIAMPolicy.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetProjectID(new.ProjectID)

		if new.Etag != "" {
			create.SetEtag(new.Etag)
		}
		if new.Version != 0 {
			create.SetVersion(new.Version)
		}
		if new.BindingsJSON != nil {
			create.SetBindingsJSON(new.BindingsJSON)
		}
		if new.AuditConfigsJSON != nil {
			create.SetAuditConfigsJSON(new.AuditConfigsJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new IAP IAM policy history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted IAP IAM policy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entiap.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPIAPIAMPolicy.Query().
		Where(
			bronzehistorygcpiapiampolicy.ResourceID(resourceID),
			bronzehistorygcpiapiampolicy.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entiap.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current IAP IAM policy history: %w", err)
	}

	err = tx.BronzeHistoryGCPIAPIAMPolicy.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close IAP IAM policy history: %w", err)
	}

	return nil
}
