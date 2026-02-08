package connector

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpvpcaccessconnector"
)

// HistoryService manages VPC Access connector history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new connector.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ConnectorData, now time.Time) error {
	create := tx.BronzeHistoryGCPVPCAccessConnector.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetProjectID(data.ProjectID)

	if data.Network != "" {
		create.SetNetwork(data.Network)
	}
	if data.IpCidrRange != "" {
		create.SetIPCidrRange(data.IpCidrRange)
	}
	if data.State != "" {
		create.SetState(data.State)
	}
	if data.MinThroughput != 0 {
		create.SetMinThroughput(data.MinThroughput)
	}
	if data.MaxThroughput != 0 {
		create.SetMaxThroughput(data.MaxThroughput)
	}
	if data.MinInstances != 0 {
		create.SetMinInstances(data.MinInstances)
	}
	if data.MaxInstances != 0 {
		create.SetMaxInstances(data.MaxInstances)
	}
	if data.MachineType != "" {
		create.SetMachineType(data.MachineType)
	}
	if data.Region != "" {
		create.SetRegion(data.Region)
	}
	if data.SubnetJSON != nil {
		create.SetSubnetJSON(data.SubnetJSON)
	}
	if data.ConnectedProjectsJSON != nil {
		create.SetConnectedProjectsJSON(data.ConnectedProjectsJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPVPCAccessConnector, new *ConnectorData, diff *ConnectorDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	// Close old history
	_, err := tx.BronzeHistoryGCPVPCAccessConnector.Update().
		Where(
			bronzehistorygcpvpcaccessconnector.ResourceID(old.ID),
			bronzehistorygcpvpcaccessconnector.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	// Create new history
	return h.CreateHistory(ctx, tx, new, now)
}

// CloseHistory closes history records for a deleted connector.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPVPCAccessConnector.Update().
		Where(
			bronzehistorygcpvpcaccessconnector.ResourceID(resourceID),
			bronzehistorygcpvpcaccessconnector.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil // No history to close
	}
	return err
}
