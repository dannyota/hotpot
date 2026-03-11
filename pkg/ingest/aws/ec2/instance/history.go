package instance

import (
	"context"
	"fmt"
	"time"

	entec2 "danny.vn/hotpot/pkg/storage/ent/aws/ec2"
	"danny.vn/hotpot/pkg/storage/ent/aws/ec2/bronzehistoryawsec2instance"
	"danny.vn/hotpot/pkg/storage/ent/aws/ec2/bronzehistoryawsec2instancetag"
)

// HistoryService handles history tracking for instances.
type HistoryService struct {
	entClient *entec2.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entec2.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new instance and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entec2.Tx, data *InstanceData, now time.Time) error {
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

	return h.createTagsHistory(ctx, tx, instHist.ID, data.Tags, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entec2.Tx, old *entec2.BronzeAWSEC2Instance, new *InstanceData, diff *InstanceDiff, now time.Time) error {
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
		if err := h.closeTagsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return fmt.Errorf("failed to close tags history: %w", err)
		}
		return h.createTagsHistory(ctx, tx, instHist.ID, new.Tags, now)
	}

	// Instance unchanged, check tags
	if diff.TagsDiff.Changed {
		if err := h.updateTagsHistory(ctx, tx, currentHist.ID, new.Tags, now); err != nil {
			return err
		}
	}

	return nil
}

// CloseHistory closes history records for a deleted instance.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entec2.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryAWSEC2Instance.Query().
		Where(
			bronzehistoryawsec2instance.ResourceID(resourceID),
			bronzehistoryawsec2instance.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entec2.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current instance history: %w", err)
	}

	if err := tx.BronzeHistoryAWSEC2Instance.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close instance history: %w", err)
	}

	return h.closeTagsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createTagsHistory(ctx context.Context, tx *entec2.Tx, instanceHistoryID uint, tags []TagData, now time.Time) error {
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

func (h *HistoryService) closeTagsHistory(ctx context.Context, tx *entec2.Tx, instanceHistoryID uint, now time.Time) error {
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

func (h *HistoryService) updateTagsHistory(ctx context.Context, tx *entec2.Tx, instanceHistoryID uint, tags []TagData, now time.Time) error {
	if err := h.closeTagsHistory(ctx, tx, instanceHistoryID, now); err != nil {
		return err
	}
	return h.createTagsHistory(ctx, tx, instanceHistoryID, tags, now)
}
