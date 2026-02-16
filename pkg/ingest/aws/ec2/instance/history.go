package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistoryawsec2instance"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistoryawsec2instancetag"
)

// HistoryService handles history tracking for instances.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new instance and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InstanceData, now time.Time) error {
	instHistCreate := tx.BronzeHistoryAWSEC2Instance.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetInstanceType(data.InstanceType).
		SetState(data.State).
		SetVpcID(data.VpcID).
		SetSubnetID(data.SubnetID).
		SetPrivateIPAddress(data.PrivateIPAddress).
		SetPublicIPAddress(data.PublicIPAddress).
		SetAmiID(data.AmiID).
		SetKeyName(data.KeyName).
		SetPlatform(data.Platform).
		SetArchitecture(data.Architecture).
		SetAccountID(data.AccountID).
		SetRegion(data.Region)

	if data.LaunchTime != nil {
		instHistCreate.SetLaunchTime(*data.LaunchTime)
	}
	if data.SecurityGroupJSON != nil {
		instHistCreate.SetSecurityGroupsJSON(data.SecurityGroupJSON)
	}

	instHist, err := instHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create instance history: %w", err)
	}

	return h.createTagsHistory(ctx, tx, instHist.HistoryID, data.Tags, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeAWSEC2Instance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryAWSEC2Instance.Query().
		Where(
			bronzehistoryawsec2instance.ResourceID(old.ID),
			bronzehistoryawsec2instance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	if diff.IsChanged {
		// Close old instance history
		if err := tx.BronzeHistoryAWSEC2Instance.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close instance history: %w", err)
		}

		// Create new instance history
		instHistCreate := tx.BronzeHistoryAWSEC2Instance.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetInstanceType(new.InstanceType).
			SetState(new.State).
			SetVpcID(new.VpcID).
			SetSubnetID(new.SubnetID).
			SetPrivateIPAddress(new.PrivateIPAddress).
			SetPublicIPAddress(new.PublicIPAddress).
			SetAmiID(new.AmiID).
			SetKeyName(new.KeyName).
			SetPlatform(new.Platform).
			SetArchitecture(new.Architecture).
			SetAccountID(new.AccountID).
			SetRegion(new.Region)

		if new.LaunchTime != nil {
			instHistCreate.SetLaunchTime(*new.LaunchTime)
		}
		if new.SecurityGroupJSON != nil {
			instHistCreate.SetSecurityGroupsJSON(new.SecurityGroupJSON)
		}

		instHist, err := instHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new instance history: %w", err)
		}

		// Close all tags history and create new ones
		if err := h.closeTagsHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close tags history: %w", err)
		}
		return h.createTagsHistory(ctx, tx, instHist.HistoryID, new.Tags, now)
	}

	// Instance unchanged, check tags
	if diff.TagsDiff.Changed {
		if err := h.updateTagsHistory(ctx, tx, currentHist.HistoryID, new.Tags, now); err != nil {
			return err
		}
	}

	return nil
}

// CloseHistory closes history records for a deleted instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryAWSEC2Instance.Query().
		Where(
			bronzehistoryawsec2instance.ResourceID(resourceID),
			bronzehistoryawsec2instance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	if err := tx.BronzeHistoryAWSEC2Instance.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close instance history: %w", err)
	}

	return h.closeTagsHistory(ctx, tx, currentHist.HistoryID, now)
}

func (h *HistoryService) createTagsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, tags []TagData, now time.Time) error {
	for _, tagData := range tags {
		_, err := tx.BronzeHistoryAWSEC2InstanceTag.Create().
			SetInstanceHistoryID(instanceHistoryID).
			SetValidFrom(now).
			SetKey(tagData.Key).
			SetValue(tagData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tag history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeTagsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryAWSEC2InstanceTag.Update().
		Where(
			bronzehistoryawsec2instancetag.InstanceHistoryID(instanceHistoryID),
			bronzehistoryawsec2instancetag.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close tag history: %w", err)
	}
	return nil
}

func (h *HistoryService) updateTagsHistory(ctx context.Context, tx *ent.Tx, instanceHistoryID uint, tags []TagData, now time.Time) error {
	if err := h.closeTagsHistory(ctx, tx, instanceHistoryID, now); err != nil {
		return err
	}
	return h.createTagsHistory(ctx, tx, instanceHistoryID, tags, now)
}
