package servergroup

import (
	"context"
	"fmt"
	"time"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzehistorygreennodecomputeservergroup"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzehistorygreennodecomputeservergroupmember"
)

// HistoryService handles history tracking for server groups.
type HistoryService struct {
	entClient *entcompute.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcompute.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new server group and members.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcompute.Tx, data *ServerGroupData, now time.Time) error {
	sgHist, err := tx.BronzeHistoryGreenNodeComputeServerGroup.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetPolicyID(data.PolicyID).
		SetPolicyName(data.PolicyName).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create server group history: %w", err)
	}
	return h.createMembersHistory(ctx, tx, sgHist.ID, data.Members, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcompute.Tx, old *entcompute.BronzeGreenNodeComputeServerGroup, new *ServerGroupData, diff *ServerGroupDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeServerGroup.Query().
		Where(
			bronzehistorygreennodecomputeservergroup.ResourceID(old.ID),
			bronzehistorygreennodecomputeservergroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current server group history: %w", err)
	}

	if diff.IsChanged {
		if err := tx.BronzeHistoryGreenNodeComputeServerGroup.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close server group history: %w", err)
		}

		sgHist, err := tx.BronzeHistoryGreenNodeComputeServerGroup.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetPolicyID(new.PolicyID).
			SetPolicyName(new.PolicyName).
			SetRegion(new.Region).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create new server group history: %w", err)
		}

		if err := h.closeMembersHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createMembersHistory(ctx, tx, sgHist.ID, new.Members, now)
	}

	if diff.MembersDiff.Changed {
		if err := h.closeMembersHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createMembersHistory(ctx, tx, currentHist.ID, new.Members, now)
	}

	return nil
}

// CloseHistory closes history for a deleted server group.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcompute.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeComputeServerGroup.Query().
		Where(
			bronzehistorygreennodecomputeservergroup.ResourceID(resourceID),
			bronzehistorygreennodecomputeservergroup.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entcompute.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current server group history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeComputeServerGroup.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close server group history: %w", err)
	}

	return h.closeMembersHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createMembersHistory(ctx context.Context, tx *entcompute.Tx, sgHistoryID uint, members []MemberData, now time.Time) error {
	for _, m := range members {
		_, err := tx.BronzeHistoryGreenNodeComputeServerGroupMember.Create().
			SetServerGroupHistoryID(sgHistoryID).
			SetValidFrom(now).
			SetUUID(m.UUID).
			SetName(m.Name).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create member history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeMembersHistory(ctx context.Context, tx *entcompute.Tx, sgHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeComputeServerGroupMember.Update().
		Where(
			bronzehistorygreennodecomputeservergroupmember.ServerGroupHistoryID(sgHistoryID),
			bronzehistorygreennodecomputeservergroupmember.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close members history: %w", err)
	}
	return nil
}
