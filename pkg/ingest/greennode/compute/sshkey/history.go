package sshkey

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodecomputesshkey"
)

// HistoryService handles history tracking for SSH keys.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new SSH key.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *SSHKeyData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeComputeSSHKey.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetPubKey(data.PubKey).
		SetStatus(data.Status).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create ssh key history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeComputeSSHKey, new *SSHKeyData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeSSHKey.Query().
		Where(
			bronzehistorygreennodecomputesshkey.ResourceID(old.ID),
			bronzehistorygreennodecomputesshkey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current ssh key history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeSSHKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ssh key history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeComputeSSHKey.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetCreatedAtAPI(new.CreatedAtAPI).
		SetPubKey(new.PubKey).
		SetStatus(new.Status).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new ssh key history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted SSH key.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeSSHKey.Query().
		Where(
			bronzehistorygreennodecomputesshkey.ResourceID(resourceID),
			bronzehistorygreennodecomputesshkey.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current ssh key history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeSSHKey.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close ssh key history: %w", err)
	}
	return nil
}
