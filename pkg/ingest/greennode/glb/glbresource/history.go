package glbresource

import (
	"context"
	"fmt"
	"time"

	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb/bronzehistorygreennodeglbgloballistener"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb/bronzehistorygreennodeglbgloballoadbalancer"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb/bronzehistorygreennodeglbglobalpool"
)

// HistoryService handles history tracking for GLBs.
type HistoryService struct {
	entClient *entglb.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entglb.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new GLB and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entglb.Tx, data *GLBData, now time.Time) error {
	glbHist, err := h.createGLBHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	if err := h.createListenersHistory(ctx, tx, glbHist.ID, data.Listeners, now); err != nil {
		return err
	}
	return h.createPoolsHistory(ctx, tx, glbHist.ID, data.Pools, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entglb.Tx, old *entglb.BronzeGreenNodeGLBGlobalLoadBalancer, new *GLBData, diff *GLBDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalLoadBalancer.Query().
		Where(
			bronzehistorygreennodeglbgloballoadbalancer.ResourceID(old.ID),
			bronzehistorygreennodeglbgloballoadbalancer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current GLB history: %w", err)
	}

	if diff.IsChanged {
		// Close old history
		if err := tx.BronzeHistoryGreenNodeGLBGlobalLoadBalancer.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close GLB history: %w", err)
		}

		// Create new history
		glbHist, err := h.createGLBHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeListenersHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		if err := h.createListenersHistory(ctx, tx, glbHist.ID, new.Listeners, now); err != nil {
			return err
		}
		if err := h.closePoolsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createPoolsHistory(ctx, tx, glbHist.ID, new.Pools, now)
	}

	// GLB unchanged, check children
	if diff.ListenersDiff.Changed {
		if err := h.closeListenersHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		if err := h.createListenersHistory(ctx, tx, currentHist.ID, new.Listeners, now); err != nil {
			return err
		}
	}

	if diff.PoolsDiff.Changed {
		if err := h.closePoolsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		if err := h.createPoolsHistory(ctx, tx, currentHist.ID, new.Pools, now); err != nil {
			return err
		}
	}

	return nil
}

// CloseHistory closes history records for a deleted GLB.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entglb.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeGLBGlobalLoadBalancer.Query().
		Where(
			bronzehistorygreennodeglbgloballoadbalancer.ResourceID(resourceID),
			bronzehistorygreennodeglbgloballoadbalancer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entglb.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current GLB history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeGLBGlobalLoadBalancer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close GLB history: %w", err)
	}

	if err := h.closeListenersHistory(ctx, tx, currentHist.ID, now); err != nil {
		return err
	}
	return h.closePoolsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createGLBHistory(ctx context.Context, tx *entglb.Tx, data *GLBData, now time.Time, firstCollectedAt time.Time) (*entglb.BronzeHistoryGreenNodeGLBGlobalLoadBalancer, error) {
	create := tx.BronzeHistoryGreenNodeGLBGlobalLoadBalancer.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetStatus(data.Status).
		SetPackage(data.Package).
		SetType(data.Type).
		SetUserID(data.UserID).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetUpdatedAtAPI(data.UpdatedAtAPI).
		SetDeletedAtAPI(data.DeletedAtAPI).
		SetProjectID(data.ProjectID)

	if data.VipsJSON != nil {
		create.SetVipsJSON(data.VipsJSON)
	}
	if data.DomainsJSON != nil {
		create.SetDomainsJSON(data.DomainsJSON)
	}

	hist, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create GLB history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createListenersHistory(ctx context.Context, tx *entglb.Tx, glbHistoryID uint, listeners []GLBListenerData, now time.Time) error {
	for _, l := range listeners {
		create := tx.BronzeHistoryGreenNodeGLBGlobalListener.Create().
			SetGlbHistoryID(glbHistoryID).
			SetValidFrom(now).
			SetListenerID(l.ListenerID).
			SetName(l.Name).
			SetDescription(l.Description).
			SetProtocol(l.Protocol).
			SetPort(l.Port).
			SetGlobalPoolID(l.GlobalPoolID).
			SetTimeoutClient(l.TimeoutClient).
			SetTimeoutMember(l.TimeoutMember).
			SetTimeoutConnection(l.TimeoutConnection).
			SetAllowedCidrs(l.AllowedCidrs).
			SetStatus(l.Status).
			SetCreatedAtAPI(l.CreatedAtAPI).
			SetUpdatedAtAPI(l.UpdatedAtAPI)

		if l.Headers != nil {
			create.SetHeaders(*l.Headers)
		}
		if l.DeletedAtAPI != nil {
			create.SetDeletedAtAPI(*l.DeletedAtAPI)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("create listener history %s: %w", l.ListenerID, err)
		}
	}
	return nil
}

func (h *HistoryService) createPoolsHistory(ctx context.Context, tx *entglb.Tx, glbHistoryID uint, pools []GLBPoolData, now time.Time) error {
	for _, p := range pools {
		create := tx.BronzeHistoryGreenNodeGLBGlobalPool.Create().
			SetGlbHistoryID(glbHistoryID).
			SetValidFrom(now).
			SetPoolID(p.PoolID).
			SetName(p.Name).
			SetDescription(p.Description).
			SetAlgorithm(p.Algorithm).
			SetProtocol(p.Protocol).
			SetStatus(p.Status).
			SetCreatedAtAPI(p.CreatedAtAPI).
			SetUpdatedAtAPI(p.UpdatedAtAPI)

		if p.StickySession != nil {
			create.SetStickySession(*p.StickySession)
		}
		if p.TLSEnabled != nil {
			create.SetTLSEnabled(*p.TLSEnabled)
		}
		if p.HealthJSON != nil {
			create.SetHealthJSON(p.HealthJSON)
		}
		if p.PoolMembersJSON != nil {
			create.SetPoolMembersJSON(p.PoolMembersJSON)
		}
		if p.DeletedAtAPI != nil {
			create.SetDeletedAtAPI(*p.DeletedAtAPI)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("create pool history %s: %w", p.PoolID, err)
		}
	}
	return nil
}

func (h *HistoryService) closeListenersHistory(ctx context.Context, tx *entglb.Tx, glbHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeGLBGlobalListener.Update().
		Where(
			bronzehistorygreennodeglbgloballistener.GlbHistoryID(glbHistoryID),
			bronzehistorygreennodeglbgloballistener.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close listeners history: %w", err)
	}
	return nil
}

func (h *HistoryService) closePoolsHistory(ctx context.Context, tx *entglb.Tx, glbHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeGLBGlobalPool.Update().
		Where(
			bronzehistorygreennodeglbglobalpool.GlbHistoryID(glbHistoryID),
			bronzehistorygreennodeglbglobalpool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close pools history: %w", err)
	}
	return nil
}
