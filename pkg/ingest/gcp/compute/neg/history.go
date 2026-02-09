package neg

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputeneg"
)

type HistoryService struct {
	entClient *ent.Client
}

func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *NegData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeNeg.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetCreationTimestamp(data.CreationTimestamp).
		SetSelfLink(data.SelfLink).
		SetNetwork(data.Network).
		SetSubnetwork(data.Subnetwork).
		SetZone(data.Zone).
		SetNetworkEndpointType(data.NetworkEndpointType).
		SetDefaultPort(data.DefaultPort).
		SetSize(data.Size).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID)

	if data.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(data.AnnotationsJSON)
	}
	if data.AppEngineJSON != nil {
		create.SetAppEngineJSON(data.AppEngineJSON)
	}
	if data.CloudFunctionJSON != nil {
		create.SetCloudFunctionJSON(data.CloudFunctionJSON)
	}
	if data.CloudRunJSON != nil {
		create.SetCloudRunJSON(data.CloudRunJSON)
	}
	if data.PscDataJSON != nil {
		create.SetPscDataJSON(data.PscDataJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create NEG history: %w", err)
	}
	return nil
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeNeg, new *NegData, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPComputeNeg.Query().
		Where(
			bronzehistorygcpcomputeneg.ResourceID(old.ID),
			bronzehistorygcpcomputeneg.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current NEG history: %w", err)
	}

	// Close current history
	if err := tx.BronzeHistoryGCPComputeNeg.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close NEG history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeNeg.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetCreationTimestamp(new.CreationTimestamp).
		SetSelfLink(new.SelfLink).
		SetNetwork(new.Network).
		SetSubnetwork(new.Subnetwork).
		SetZone(new.Zone).
		SetNetworkEndpointType(new.NetworkEndpointType).
		SetDefaultPort(new.DefaultPort).
		SetSize(new.Size).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID)

	if new.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(new.AnnotationsJSON)
	}
	if new.AppEngineJSON != nil {
		create.SetAppEngineJSON(new.AppEngineJSON)
	}
	if new.CloudFunctionJSON != nil {
		create.SetCloudFunctionJSON(new.CloudFunctionJSON)
	}
	if new.CloudRunJSON != nil {
		create.SetCloudRunJSON(new.CloudRunJSON)
	}
	if new.PscDataJSON != nil {
		create.SetPscDataJSON(new.PscDataJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create new NEG history: %w", err)
	}

	return nil
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPComputeNeg.Query().
		Where(
			bronzehistorygcpcomputeneg.ResourceID(resourceID),
			bronzehistorygcpcomputeneg.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current NEG history: %w", err)
	}

	if err := tx.BronzeHistoryGCPComputeNeg.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close NEG history: %w", err)
	}

	return nil
}
