package secret

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsecretmanagersecret"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsecretmanagersecretlabel"
)

// HistoryService handles history tracking for secrets.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new secret and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, secretData *SecretData, now time.Time) error {
	hist, err := tx.BronzeHistoryGCPSecretManagerSecret.Create().
		SetResourceID(secretData.ID).
		SetValidFrom(now).
		SetCollectedAt(secretData.CollectedAt).
		SetFirstCollectedAt(secretData.CollectedAt).
		SetName(secretData.Name).
		SetCreateTime(secretData.CreateTime).
		SetEtag(secretData.Etag).
		SetReplicationJSON(secretData.ReplicationJSON).
		SetRotationJSON(secretData.RotationJSON).
		SetTopicsJSON(secretData.TopicsJSON).
		SetVersionAliasesJSON(secretData.VersionAliasesJSON).
		SetAnnotationsJSON(secretData.AnnotationsJSON).
		SetProjectID(secretData.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create secret history: %w", err)
	}

	return h.createLabelsHistory(ctx, tx, hist.HistoryID, secretData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSecretManagerSecret, new *SecretData, diff *SecretDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPSecretManagerSecret.Query().
		Where(
			bronzehistorygcpsecretmanagersecret.ResourceID(old.ID),
			bronzehistorygcpsecretmanagersecret.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current secret history: %w", err)
	}

	if diff.IsChanged {
		err = tx.BronzeHistoryGCPSecretManagerSecret.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current secret history: %w", err)
		}

		hist, err := tx.BronzeHistoryGCPSecretManagerSecret.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetCreateTime(new.CreateTime).
			SetEtag(new.Etag).
			SetReplicationJSON(new.ReplicationJSON).
			SetRotationJSON(new.RotationJSON).
			SetTopicsJSON(new.TopicsJSON).
			SetVersionAliasesJSON(new.VersionAliasesJSON).
			SetAnnotationsJSON(new.AnnotationsJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new secret history: %w", err)
		}

		if err := h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close labels history: %w", err)
		}
		return h.createLabelsHistory(ctx, tx, hist.HistoryID, new, now)
	}

	if diff.LabelDiff.Changed {
		return h.updateLabelsHistory(ctx, tx, currentHist.HistoryID, new, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted secret.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGCPSecretManagerSecret.Query().
		Where(
			bronzehistorygcpsecretmanagersecret.ResourceID(resourceID),
			bronzehistorygcpsecretmanagersecret.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current secret history: %w", err)
	}

	err = tx.BronzeHistoryGCPSecretManagerSecret.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close secret history: %w", err)
	}

	return h.closeLabelsHistory(ctx, tx, currentHist.HistoryID, now)
}

func (h *HistoryService) createLabelsHistory(ctx context.Context, tx *ent.Tx, secretHistoryID uint, secretData *SecretData, now time.Time) error {
	for _, label := range secretData.Labels {
		_, err := tx.BronzeHistoryGCPSecretManagerSecretLabel.Create().
			SetSecretHistoryID(secretHistoryID).
			SetValidFrom(now).
			SetKey(label.Key).
			SetValue(label.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeLabelsHistory(ctx context.Context, tx *ent.Tx, secretHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGCPSecretManagerSecretLabel.Update().
		Where(
			bronzehistorygcpsecretmanagersecretlabel.SecretHistoryID(secretHistoryID),
			bronzehistorygcpsecretmanagersecretlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}
	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, secretHistoryID uint, new *SecretData, now time.Time) error {
	if err := h.closeLabelsHistory(ctx, tx, secretHistoryID, now); err != nil {
		return err
	}
	return h.createLabelsHistory(ctx, tx, secretHistoryID, new, now)
}
