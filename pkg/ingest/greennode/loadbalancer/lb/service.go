package lb

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeloadbalancerlb"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeloadbalancerlistener"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeloadbalancerpool"
)

// Service handles GreenNode load balancer ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new load balancer ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of load balancer ingestion.
type IngestResult struct {
	LBCount        int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches load balancers from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	lbs, err := s.client.ListLoadBalancers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list load balancers: %w", err)
	}

	lbDataList := make([]*LBData, 0, len(lbs))
	for _, lb := range lbs {
		listeners, err := s.client.ListListenersByLBID(ctx, lb.UUID)
		if err != nil {
			return nil, fmt.Errorf("list listeners for LB %s: %w", lb.UUID, err)
		}

		pools, err := s.client.ListPoolsByLBID(ctx, lb.UUID)
		if err != nil {
			return nil, fmt.Errorf("list pools for LB %s: %w", lb.UUID, err)
		}

		data, err := ConvertLB(lb, listeners, pools, projectID, region, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert load balancer: %w", err)
		}
		lbDataList = append(lbDataList, data)
	}

	if err := s.saveLBs(ctx, lbDataList); err != nil {
		return nil, fmt.Errorf("save load balancers: %w", err)
	}

	return &IngestResult{
		LBCount:        len(lbDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveLBs(ctx context.Context, lbs []*LBData) error {
	if len(lbs) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range lbs {
		existing, err := tx.BronzeGreenNodeLoadBalancerLB.Query().
			Where(bronzegreennodeloadbalancerlb.ID(data.ID)).
			WithListeners().
			WithPools().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing LB %s: %w", data.Name, err)
		}

		diff := DiffLBData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeLoadBalancerLB.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for LB %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteLBChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for LB %s: %w", data.Name, err)
			}
		}

		var savedLB *ent.BronzeGreenNodeLoadBalancerLB
		if existing == nil {
			create := tx.BronzeGreenNodeLoadBalancerLB.Create().
				SetID(data.ID).
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
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.NodesJSON != nil {
				create.SetNodesJSON(data.NodesJSON)
			}

			savedLB, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create LB %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGreenNodeLoadBalancerLB.UpdateOneID(data.ID).
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
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.NodesJSON != nil {
				update.SetNodesJSON(data.NodesJSON)
			}

			savedLB, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update LB %s: %w", data.Name, err)
			}
		}

		if err := s.createLBChildren(ctx, tx, savedLB, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for LB %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for LB %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for LB %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteLBChildren(ctx context.Context, tx *ent.Tx, lbID string) error {
	_, err := tx.BronzeGreenNodeLoadBalancerListener.Delete().
		Where(bronzegreennodeloadbalancerlistener.HasLbWith(bronzegreennodeloadbalancerlb.ID(lbID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete listeners: %w", err)
	}

	_, err = tx.BronzeGreenNodeLoadBalancerPool.Delete().
		Where(bronzegreennodeloadbalancerpool.HasLbWith(bronzegreennodeloadbalancerlb.ID(lbID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete pools: %w", err)
	}

	return nil
}

func (s *Service) createLBChildren(ctx context.Context, tx *ent.Tx, lb *ent.BronzeGreenNodeLoadBalancerLB, data *LBData) error {
	for _, l := range data.Listeners {
		create := tx.BronzeGreenNodeLoadBalancerListener.Create().
			SetLb(lb).
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
			return fmt.Errorf("create listener %s: %w", l.ListenerID, err)
		}
	}

	for _, p := range data.Pools {
		create := tx.BronzeGreenNodeLoadBalancerPool.Create().
			SetLb(lb).
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
			return fmt.Errorf("create pool %s: %w", p.PoolID, err)
		}
	}

	return nil
}

// DeleteStaleLBs removes load balancers not collected in the latest run for the given region.
func (s *Service) DeleteStaleLBs(ctx context.Context, projectID, region string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeLoadBalancerLB.Query().
		Where(
			bronzegreennodeloadbalancerlb.ProjectID(projectID),
			bronzegreennodeloadbalancerlb.Region(region),
			bronzegreennodeloadbalancerlb.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale load balancers: %w", err)
	}

	for _, lb := range stale {
		if err := s.history.CloseHistory(ctx, tx, lb.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for LB %s: %w", lb.ID, err)
		}
		if err := s.deleteLBChildren(ctx, tx, lb.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for LB %s: %w", lb.ID, err)
		}
		if err := tx.BronzeGreenNodeLoadBalancerLB.DeleteOneID(lb.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete LB %s: %w", lb.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
