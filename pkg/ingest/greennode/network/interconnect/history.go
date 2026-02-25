package interconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodenetworkinterconnect"
)

// HistoryService handles history tracking for interconnects.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new interconnect.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *InterconnectData, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkInterconnect.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetStatus(data.Status).
		SetEnableGw2(data.EnableGw2).
		SetCircuitID(data.CircuitID).
		SetGw01IP(data.Gw01IP).
		SetGw02IP(data.Gw02IP).
		SetGwVip(data.GwVIP).
		SetRemoteGw01IP(data.RemoteGw01IP).
		SetRemoteGw02IP(data.RemoteGw02IP).
		SetPackageID(data.PackageID).
		SetTypeID(data.TypeID).
		SetTypeName(data.TypeName).
		SetCreatedAt(data.CreatedAt).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create interconnect history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeNetworkInterconnect, new *InterconnectData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkInterconnect.Query().
		Where(
			bronzehistorygreennodenetworkinterconnect.ResourceID(old.ID),
			bronzehistorygreennodenetworkinterconnect.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current interconnect history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkInterconnect.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close interconnect history: %w", err)
	}

	_, err = tx.BronzeHistoryGreenNodeNetworkInterconnect.Create().
		SetResourceID(new.UUID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetDescription(new.Description).
		SetStatus(new.Status).
		SetEnableGw2(new.EnableGw2).
		SetCircuitID(new.CircuitID).
		SetGw01IP(new.Gw01IP).
		SetGw02IP(new.Gw02IP).
		SetGwVip(new.GwVIP).
		SetRemoteGw01IP(new.RemoteGw01IP).
		SetRemoteGw02IP(new.RemoteGw02IP).
		SetPackageID(new.PackageID).
		SetTypeID(new.TypeID).
		SetTypeName(new.TypeName).
		SetCreatedAt(new.CreatedAt).
		SetRegion(new.Region).
		SetProjectID(new.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new interconnect history: %w", err)
	}
	return nil
}

// CloseHistory closes history for a deleted interconnect.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkInterconnect.Query().
		Where(
			bronzehistorygreennodenetworkinterconnect.ResourceID(resourceID),
			bronzehistorygreennodenetworkinterconnect.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current interconnect history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkInterconnect.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close interconnect history: %w", err)
	}
	return nil
}
