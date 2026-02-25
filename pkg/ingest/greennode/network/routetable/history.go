package routetable

import (
	"context"
	"fmt"
	"time"

	entnet "github.com/dannyota/hotpot/pkg/storage/ent/greennode/network"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzehistorygreennodenetworkroutetable"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/network/bronzehistorygreennodenetworkroutetableroute"
)

// HistoryService handles history tracking for route tables.
type HistoryService struct {
	entClient *entnet.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entnet.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new route table and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entnet.Tx, data *RouteTableData, now time.Time) error {
	rtHist, err := h.createRouteTableHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	return h.createRoutesHistory(ctx, tx, rtHist.ID, data.Routes, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entnet.Tx, old *entnet.BronzeGreenNodeNetworkRouteTable, new *RouteTableData, diff *RouteTableDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkRouteTable.Query().
		Where(
			bronzehistorygreennodenetworkroutetable.ResourceID(old.ID),
			bronzehistorygreennodenetworkroutetable.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current route table history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeNetworkRouteTable.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close route table history: %w", err)
		}

		// Create new history
		rtHist, err := h.createRouteTableHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeRoutesHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRoutesHistory(ctx, tx, rtHist.ID, new.Routes, now)
	}

	// Route table unchanged, check children
	if diff.RoutesDiff.Changed {
		if err := h.closeRoutesHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createRoutesHistory(ctx, tx, currentHist.ID, new.Routes, now)
	}

	return nil
}

// CloseHistory closes history records for a deleted route table.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entnet.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeNetworkRouteTable.Query().
		Where(
			bronzehistorygreennodenetworkroutetable.ResourceID(resourceID),
			bronzehistorygreennodenetworkroutetable.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entnet.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current route table history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeNetworkRouteTable.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close route table history: %w", err)
	}

	return h.closeRoutesHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createRouteTableHistory(ctx context.Context, tx *entnet.Tx, data *RouteTableData, now time.Time, firstCollectedAt time.Time) (*entnet.BronzeHistoryGreenNodeNetworkRouteTable, error) {
	hist, err := tx.BronzeHistoryGreenNodeNetworkRouteTable.Create().
		SetResourceID(data.UUID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetName(data.Name).
		SetStatus(data.Status).
		SetNetworkID(data.NetworkID).
		SetCreatedAt(data.CreatedAt).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create route table history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createRoutesHistory(ctx context.Context, tx *entnet.Tx, routeTableHistoryID uint, routes []RouteData, now time.Time) error {
	for _, r := range routes {
		_, err := tx.BronzeHistoryGreenNodeNetworkRouteTableRoute.Create().
			SetRouteTableHistoryID(routeTableHistoryID).
			SetValidFrom(now).
			SetRouteID(r.RouteID).
			SetRouteTableID(r.RouteTableID).
			SetRoutingType(r.RoutingType).
			SetDestinationCidrBlock(r.DestinationCidrBlock).
			SetTarget(r.Target).
			SetStatus(r.Status).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create route history: %w", err)
		}
	}
	return nil
}

func (h *HistoryService) closeRoutesHistory(ctx context.Context, tx *entnet.Tx, routeTableHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeNetworkRouteTableRoute.Update().
		Where(
			bronzehistorygreennodenetworkroutetableroute.RouteTableHistoryID(routeTableHistoryID),
			bronzehistorygreennodenetworkroutetableroute.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close routes history: %w", err)
	}
	return nil
}
