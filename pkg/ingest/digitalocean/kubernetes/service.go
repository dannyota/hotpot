package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedokubernetescluster"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedokubernetesnodepool"
)

// Service handles DigitalOcean Kubernetes ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	clusterHistory *ClusterHistoryService
	npHistory      *NodePoolHistoryService
}

// NewService creates a new Kubernetes ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:         client,
		entClient:      entClient,
		clusterHistory: NewClusterHistoryService(entClient),
		npHistory:      NewNodePoolHistoryService(entClient),
	}
}

// IngestClustersResult contains the result of Kubernetes cluster ingestion.
type IngestClustersResult struct {
	ClusterCount   int
	CollectedAt    time.Time
	DurationMillis int64
	ClusterIDs     []string
}

// IngestClusters fetches all Kubernetes clusters from DigitalOcean and saves them.
func (s *Service) IngestClusters(ctx context.Context, heartbeat func()) (*IngestClustersResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiClusters, err := s.client.ListAllClusters(ctx)
	if err != nil {
		return nil, fmt.Errorf("list kubernetes clusters: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allClusters []*ClusterData
	var clusterIDs []string
	for _, v := range apiClusters {
		allClusters = append(allClusters, ConvertCluster(v, collectedAt))
		clusterIDs = append(clusterIDs, v.ID)
	}

	if err := s.saveClusters(ctx, allClusters); err != nil {
		return nil, fmt.Errorf("save kubernetes clusters: %w", err)
	}

	return &IngestClustersResult{
		ClusterCount:   len(allClusters),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
		ClusterIDs:     clusterIDs,
	}, nil
}

func (s *Service) saveClusters(ctx context.Context, clusters []*ClusterData) error {
	if len(clusters) == 0 {
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

	for _, data := range clusters {
		existing, err := tx.BronzeDOKubernetesCluster.Query().
			Where(bronzedokubernetescluster.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing KubernetesCluster %s: %w", data.ResourceID, err)
		}

		diff := DiffClusterData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOKubernetesCluster.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for KubernetesCluster %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOKubernetesCluster.Create().
				SetID(data.ResourceID).
				SetName(data.Name).
				SetRegionSlug(data.RegionSlug).
				SetVersionSlug(data.VersionSlug).
				SetClusterSubnet(data.ClusterSubnet).
				SetServiceSubnet(data.ServiceSubnet).
				SetIpv4(data.IPv4).
				SetEndpoint(data.Endpoint).
				SetVpcUUID(data.VPCUUID).
				SetHa(data.HA).
				SetAutoUpgrade(data.AutoUpgrade).
				SetSurgeUpgrade(data.SurgeUpgrade).
				SetRegistryEnabled(data.RegistryEnabled).
				SetStatusState(data.StatusState).
				SetStatusMessage(data.StatusMessage).
				SetTagsJSON(data.TagsJSON).
				SetMaintenancePolicyJSON(data.MaintenancePolicyJSON).
				SetControlPlaneFirewallJSON(data.ControlPlaneFirewallJSON).
				SetAutoscalerConfigJSON(data.AutoscalerConfigJSON).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetNillableAPIUpdatedAt(data.APIUpdatedAt).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create KubernetesCluster %s: %w", data.ResourceID, err)
			}

			if err := s.clusterHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for KubernetesCluster %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOKubernetesCluster.UpdateOneID(data.ResourceID).
				SetName(data.Name).
				SetRegionSlug(data.RegionSlug).
				SetVersionSlug(data.VersionSlug).
				SetClusterSubnet(data.ClusterSubnet).
				SetServiceSubnet(data.ServiceSubnet).
				SetIpv4(data.IPv4).
				SetEndpoint(data.Endpoint).
				SetVpcUUID(data.VPCUUID).
				SetHa(data.HA).
				SetAutoUpgrade(data.AutoUpgrade).
				SetSurgeUpgrade(data.SurgeUpgrade).
				SetRegistryEnabled(data.RegistryEnabled).
				SetStatusState(data.StatusState).
				SetStatusMessage(data.StatusMessage).
				SetTagsJSON(data.TagsJSON).
				SetMaintenancePolicyJSON(data.MaintenancePolicyJSON).
				SetControlPlaneFirewallJSON(data.ControlPlaneFirewallJSON).
				SetAutoscalerConfigJSON(data.AutoscalerConfigJSON).
				SetNillableAPICreatedAt(data.APICreatedAt).
				SetNillableAPIUpdatedAt(data.APIUpdatedAt).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update KubernetesCluster %s: %w", data.ResourceID, err)
			}

			if err := s.clusterHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for KubernetesCluster %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleClusters removes Kubernetes clusters that were not collected in the latest run.
func (s *Service) DeleteStaleClusters(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeDOKubernetesCluster.Query().
		Where(bronzedokubernetescluster.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, d := range stale {
		if err := s.clusterHistory.CloseHistory(ctx, tx, d.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for KubernetesCluster %s: %w", d.ID, err)
		}

		if err := tx.BronzeDOKubernetesCluster.DeleteOne(d).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete KubernetesCluster %s: %w", d.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// IngestNodePoolsResult contains the result of Kubernetes node pool ingestion.
type IngestNodePoolsResult struct {
	NodePoolCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// IngestNodePools fetches all node pools for given clusters and saves them.
func (s *Service) IngestNodePools(ctx context.Context, clusterIDs []string, heartbeat func()) (*IngestNodePoolsResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allNodePools []*NodePoolData

	for _, clusterID := range clusterIDs {
		apiPools, err := s.client.ListAllNodePools(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("list node pools for cluster %s: %w", clusterID, err)
		}
		for _, v := range apiPools {
			allNodePools = append(allNodePools, ConvertNodePool(v, clusterID, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}
	}

	if err := s.saveNodePools(ctx, allNodePools); err != nil {
		return nil, fmt.Errorf("save node pools: %w", err)
	}

	return &IngestNodePoolsResult{
		NodePoolCount:  len(allNodePools),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveNodePools(ctx context.Context, pools []*NodePoolData) error {
	if len(pools) == 0 {
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

	for _, data := range pools {
		existing, err := tx.BronzeDOKubernetesNodePool.Query().
			Where(bronzedokubernetesnodepool.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing KubernetesNodePool %s: %w", data.ResourceID, err)
		}

		diff := DiffNodePoolData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOKubernetesNodePool.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for KubernetesNodePool %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOKubernetesNodePool.Create().
				SetID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetNodePoolID(data.NodePoolID).
				SetName(data.Name).
				SetSize(data.Size).
				SetCount(data.Count).
				SetAutoScale(data.AutoScale).
				SetMinNodes(data.MinNodes).
				SetMaxNodes(data.MaxNodes).
				SetTagsJSON(data.TagsJSON).
				SetLabelsJSON(data.LabelsJSON).
				SetTaintsJSON(data.TaintsJSON).
				SetNodesJSON(data.NodesJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create KubernetesNodePool %s: %w", data.ResourceID, err)
			}
			if err := s.npHistory.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for KubernetesNodePool %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOKubernetesNodePool.UpdateOneID(data.ResourceID).
				SetClusterID(data.ClusterID).
				SetNodePoolID(data.NodePoolID).
				SetName(data.Name).
				SetSize(data.Size).
				SetCount(data.Count).
				SetAutoScale(data.AutoScale).
				SetMinNodes(data.MinNodes).
				SetMaxNodes(data.MaxNodes).
				SetTagsJSON(data.TagsJSON).
				SetLabelsJSON(data.LabelsJSON).
				SetTaintsJSON(data.TaintsJSON).
				SetNodesJSON(data.NodesJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update KubernetesNodePool %s: %w", data.ResourceID, err)
			}
			if err := s.npHistory.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for KubernetesNodePool %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStaleNodePools removes node pools not collected in the latest run.
func (s *Service) DeleteStaleNodePools(ctx context.Context, collectedAt time.Time) error {
	return s.deleteStale(ctx, collectedAt, "KubernetesNodePool",
		func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error) {
			stale, err := tx.BronzeDOKubernetesNodePool.Query().
				Where(bronzedokubernetesnodepool.CollectedAtLT(collectedAt)).
				All(ctx)
			if err != nil {
				return nil, err
			}
			result := make([]*staleResource, len(stale))
			for i, r := range stale {
				result[i] = &staleResource{id: r.ID, delete: func(ctx context.Context) error {
					return tx.BronzeDOKubernetesNodePool.DeleteOne(r).Exec(ctx)
				}}
			}
			return result, nil
		},
		func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error {
			return s.npHistory.CloseHistory(ctx, tx, id, now)
		},
	)
}

type staleResource struct {
	id     string
	delete func(ctx context.Context) error
}

type queryStaleFunc func(ctx context.Context, tx *ent.Tx, collectedAt time.Time) ([]*staleResource, error)
type closeHistoryFunc func(ctx context.Context, tx *ent.Tx, id string, now time.Time) error

func (s *Service) deleteStale(ctx context.Context, collectedAt time.Time, typeName string, queryFn queryStaleFunc, closeFn closeHistoryFunc) error {
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

	stale, err := queryFn(ctx, tx, collectedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, r := range stale {
		if err := closeFn(ctx, tx, r.id, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for %s %s: %w", typeName, r.id, err)
		}
		if err := r.delete(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete %s %s: %w", typeName, r.id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
