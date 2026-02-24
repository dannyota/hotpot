package glbresource

import (
	"context"
	"fmt"
	"time"

	glbv1 "danny.vn/greennode/services/glb/v1"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeglbgloballistener"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeglbgloballoadbalancer"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodeglbglobalpool"
)

// Service handles GreenNode GLB ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new GLB ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of GLB ingestion.
type IngestResult struct {
	GLBCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches GLBs from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	glbs, err := s.client.ListGlobalLoadBalancers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list global load balancers: %w", err)
	}

	glbDataList := make([]*GLBData, 0, len(glbs))
	for _, glb := range glbs {
		data, err := ConvertGLB(glb, projectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert GLB: %w", err)
		}

		// Fetch listeners
		listeners, err := s.client.ListGlobalListeners(ctx, glb.ID)
		if err != nil {
			return nil, fmt.Errorf("list listeners for GLB %s: %w", glb.ID, err)
		}
		data.Listeners = ConvertListeners(listeners)

		// Fetch pools
		pools, err := s.client.ListGlobalPools(ctx, glb.ID)
		if err != nil {
			return nil, fmt.Errorf("list pools for GLB %s: %w", glb.ID, err)
		}

		// Fetch pool members for each pool
		poolMembers := make(map[string][]*glbv1.GlobalPoolMember)
		for _, pool := range pools {
			members, err := s.client.ListGlobalPoolMembers(ctx, glb.ID, pool.ID)
			if err != nil {
				return nil, fmt.Errorf("list pool members for pool %s: %w", pool.ID, err)
			}
			poolMembers[pool.ID] = members
		}

		convertedPools, err := ConvertPools(pools, poolMembers)
		if err != nil {
			return nil, fmt.Errorf("convert pools for GLB %s: %w", glb.ID, err)
		}
		data.Pools = convertedPools

		glbDataList = append(glbDataList, data)
	}

	if err := s.saveGLBs(ctx, glbDataList); err != nil {
		return nil, fmt.Errorf("save GLBs: %w", err)
	}

	return &IngestResult{
		GLBCount:       len(glbDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveGLBs(ctx context.Context, glbs []*GLBData) error {
	if len(glbs) == 0 {
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

	for _, data := range glbs {
		existing, err := tx.BronzeGreenNodeGLBGlobalLoadBalancer.Query().
			Where(bronzegreennodeglbgloballoadbalancer.ID(data.ID)).
			WithListeners().
			WithPools().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing GLB %s: %w", data.ID, err)
		}

		diff := DiffGLBData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeGLBGlobalLoadBalancer.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for GLB %s: %w", data.ID, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteGLBChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for GLB %s: %w", data.ID, err)
			}
		}

		var savedGLB *ent.BronzeGreenNodeGLBGlobalLoadBalancer
		if existing == nil {
			create := tx.BronzeGreenNodeGLBGlobalLoadBalancer.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetPackage(data.Package).
				SetType(data.Type).
				SetUserID(data.UserID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetDeletedAtAPI(data.DeletedAtAPI).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.VipsJSON != nil {
				create.SetVipsJSON(data.VipsJSON)
			}
			if data.DomainsJSON != nil {
				create.SetDomainsJSON(data.DomainsJSON)
			}

			savedGLB, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create GLB %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGreenNodeGLBGlobalLoadBalancer.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetPackage(data.Package).
				SetType(data.Type).
				SetUserID(data.UserID).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetDeletedAtAPI(data.DeletedAtAPI).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if data.VipsJSON != nil {
				update.SetVipsJSON(data.VipsJSON)
			}
			if data.DomainsJSON != nil {
				update.SetDomainsJSON(data.DomainsJSON)
			}

			savedGLB, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update GLB %s: %w", data.ID, err)
			}
		}

		if err := s.createGLBChildren(ctx, tx, savedGLB, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for GLB %s: %w", data.ID, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for GLB %s: %w", data.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for GLB %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteGLBChildren(ctx context.Context, tx *ent.Tx, glbID string) error {
	_, err := tx.BronzeGreenNodeGLBGlobalListener.Delete().
		Where(bronzegreennodeglbgloballistener.HasGlbWith(bronzegreennodeglbgloballoadbalancer.ID(glbID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete listeners: %w", err)
	}

	_, err = tx.BronzeGreenNodeGLBGlobalPool.Delete().
		Where(bronzegreennodeglbglobalpool.HasGlbWith(bronzegreennodeglbgloballoadbalancer.ID(glbID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete pools: %w", err)
	}

	return nil
}

func (s *Service) createGLBChildren(ctx context.Context, tx *ent.Tx, glb *ent.BronzeGreenNodeGLBGlobalLoadBalancer, data *GLBData) error {
	for _, l := range data.Listeners {
		create := tx.BronzeGreenNodeGLBGlobalListener.Create().
			SetGlb(glb).
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
			return fmt.Errorf("create listener %s: %w", l.ListenerID, err)
		}
	}

	for _, p := range data.Pools {
		create := tx.BronzeGreenNodeGLBGlobalPool.Create().
			SetGlb(glb).
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
			return fmt.Errorf("create pool %s: %w", p.PoolID, err)
		}
	}

	return nil
}

// DeleteStaleGLBs removes GLBs not collected in the latest run.
func (s *Service) DeleteStaleGLBs(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeGLBGlobalLoadBalancer.Query().
		Where(
			bronzegreennodeglbgloballoadbalancer.ProjectID(projectID),
			bronzegreennodeglbgloballoadbalancer.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale GLBs: %w", err)
	}

	for _, glb := range stale {
		if err := s.history.CloseHistory(ctx, tx, glb.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for GLB %s: %w", glb.ID, err)
		}
		if err := s.deleteGLBChildren(ctx, tx, glb.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for GLB %s: %w", glb.ID, err)
		}
		if err := tx.BronzeGreenNodeGLBGlobalLoadBalancer.DeleteOneID(glb.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete GLB %s: %w", glb.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
