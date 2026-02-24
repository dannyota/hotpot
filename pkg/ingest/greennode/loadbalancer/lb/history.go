package lb

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodeloadbalancerlb"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodeloadbalancerlistener"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygreennodeloadbalancerpool"
)

// HistoryService handles history tracking for load balancers.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new load balancer and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *LBData, now time.Time) error {
	lbHist, err := h.createLBHistory(ctx, tx, data, now, data.CollectedAt)
	if err != nil {
		return err
	}
	if err := h.createListenersHistory(ctx, tx, lbHist.ID, data.Listeners, now); err != nil {
		return err
	}
	return h.createPoolsHistory(ctx, tx, lbHist.ID, data.Pools, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGreenNodeLoadBalancerLB, new *LBData, diff *LBDiff, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerLB.Query().
		Where(
			bronzehistorygreennodeloadbalancerlb.ResourceID(old.ID),
			bronzehistorygreennodeloadbalancerlb.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current LB history: %w", err)
	}

	if diff.IsChanged {
		// Close old parent history
		if err := tx.BronzeHistoryGreenNodeLoadBalancerLB.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("close LB history: %w", err)
		}

		// Create new parent history
		lbHist, err := h.createLBHistory(ctx, tx, new, now, old.FirstCollectedAt)
		if err != nil {
			return err
		}

		// Close and recreate all children
		if err := h.closeListenersHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		if err := h.createListenersHistory(ctx, tx, lbHist.ID, new.Listeners, now); err != nil {
			return err
		}
		if err := h.closePoolsHistory(ctx, tx, currentHist.ID, now); err != nil {
			return err
		}
		return h.createPoolsHistory(ctx, tx, lbHist.ID, new.Pools, now)
	}

	// Parent unchanged, check children independently
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

// CloseHistory closes history records for a deleted load balancer.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryGreenNodeLoadBalancerLB.Query().
		Where(
			bronzehistorygreennodeloadbalancerlb.ResourceID(resourceID),
			bronzehistorygreennodeloadbalancerlb.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current LB history: %w", err)
	}

	if err := tx.BronzeHistoryGreenNodeLoadBalancerLB.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close LB history: %w", err)
	}

	if err := h.closeListenersHistory(ctx, tx, currentHist.ID, now); err != nil {
		return err
	}
	return h.closePoolsHistory(ctx, tx, currentHist.ID, now)
}

func (h *HistoryService) createLBHistory(ctx context.Context, tx *ent.Tx, data *LBData, now time.Time, firstCollectedAt time.Time) (*ent.BronzeHistoryGreenNodeLoadBalancerLB, error) {
	create := tx.BronzeHistoryGreenNodeLoadBalancerLB.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(firstCollectedAt).
		SetName(data.Name).
		SetDisplayStatus(data.DisplayStatus).
		SetAddress(data.Address).
		SetPrivateSubnetID(data.PrivateSubnetID).
		SetPrivateSubnetCidr(data.PrivateSubnetCidr).
		SetType(data.Type).
		SetDisplayType(data.DisplayType).
		SetLoadBalancerSchema(data.LoadBalancerSchema).
		SetPackageID(data.PackageID).
		SetDescription(data.Description).
		SetLocation(data.Location).
		SetCreatedAtAPI(data.CreatedAtAPI).
		SetUpdatedAtAPI(data.UpdatedAtAPI).
		SetProgressStatus(data.ProgressStatus).
		SetStatus(data.Status).
		SetBackendSubnetID(data.BackendSubnetID).
		SetInternal(data.Internal).
		SetAutoScalable(data.AutoScalable).
		SetZoneID(data.ZoneID).
		SetMinSize(data.MinSize).
		SetMaxSize(data.MaxSize).
		SetTotalNodes(data.TotalNodes).
		SetRegion(data.Region).
		SetProjectID(data.ProjectID)

	if data.NodesJSON != nil {
		create.SetNodesJSON(data.NodesJSON)
	}

	hist, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create LB history: %w", err)
	}
	return hist, nil
}

func (h *HistoryService) createListenersHistory(ctx context.Context, tx *ent.Tx, lbHistoryID uint, listeners []ListenerData, now time.Time) error {
	for _, l := range listeners {
		create := tx.BronzeHistoryGreenNodeLoadBalancerListener.Create().
			SetLbHistoryID(lbHistoryID).
			SetValidFrom(now).
			SetListenerID(l.ListenerID).
			SetName(l.Name).
			SetDescription(l.Description).
			SetProtocol(l.Protocol).
			SetProtocolPort(l.ProtocolPort).
			SetConnectionLimit(l.ConnectionLimit).
			SetDefaultPoolID(l.DefaultPoolID).
			SetDefaultPoolName(l.DefaultPoolName).
			SetTimeoutClient(l.TimeoutClient).
			SetTimeoutMember(l.TimeoutMember).
			SetTimeoutConnection(l.TimeoutConnection).
			SetAllowedCidrs(l.AllowedCidrs).
			SetDisplayStatus(l.DisplayStatus).
			SetCreatedAtAPI(l.CreatedAtAPI).
			SetUpdatedAtAPI(l.UpdatedAtAPI).
			SetProgressStatus(l.ProgressStatus)

		if l.CertificateAuthoritiesJSON != nil {
			create.SetCertificateAuthoritiesJSON(l.CertificateAuthoritiesJSON)
		}
		if l.DefaultCertificateAuthority != nil {
			create.SetDefaultCertificateAuthority(*l.DefaultCertificateAuthority)
		}
		if l.ClientCertificateAuthentication != nil {
			create.SetClientCertificateAuthentication(*l.ClientCertificateAuthentication)
		}
		if l.InsertHeadersJSON != nil {
			create.SetInsertHeadersJSON(l.InsertHeadersJSON)
		}
		if l.PoliciesJSON != nil {
			create.SetPoliciesJSON(l.PoliciesJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("create listener history %s: %w", l.ListenerID, err)
		}
	}
	return nil
}

func (h *HistoryService) createPoolsHistory(ctx context.Context, tx *ent.Tx, lbHistoryID uint, pools []PoolData, now time.Time) error {
	for _, p := range pools {
		create := tx.BronzeHistoryGreenNodeLoadBalancerPool.Create().
			SetLbHistoryID(lbHistoryID).
			SetValidFrom(now).
			SetPoolID(p.PoolID).
			SetName(p.Name).
			SetProtocol(p.Protocol).
			SetDescription(p.Description).
			SetLoadBalanceMethod(p.LoadBalanceMethod).
			SetStatus(p.Status).
			SetStickiness(p.Stickiness).
			SetTLSEncryption(p.TLSEncryption)

		if p.MembersJSON != nil {
			create.SetMembersJSON(p.MembersJSON)
		}
		if p.HealthMonitorJSON != nil {
			create.SetHealthMonitorJSON(p.HealthMonitorJSON)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("create pool history %s: %w", p.PoolID, err)
		}
	}
	return nil
}

func (h *HistoryService) closeListenersHistory(ctx context.Context, tx *ent.Tx, lbHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeLoadBalancerListener.Update().
		Where(
			bronzehistorygreennodeloadbalancerlistener.LbHistoryID(lbHistoryID),
			bronzehistorygreennodeloadbalancerlistener.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close listeners history: %w", err)
	}
	return nil
}

func (h *HistoryService) closePoolsHistory(ctx context.Context, tx *ent.Tx, lbHistoryID uint, now time.Time) error {
	_, err := tx.BronzeHistoryGreenNodeLoadBalancerPool.Update().
		Where(
			bronzehistorygreennodeloadbalancerpool.LbHistoryID(lbHistoryID),
			bronzehistorygreennodeloadbalancerpool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close pools history: %w", err)
	}
	return nil
}
