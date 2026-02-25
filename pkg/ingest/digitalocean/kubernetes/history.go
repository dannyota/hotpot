package kubernetes

import (
	"context"
	"fmt"
	"time"

	entdo "github.com/dannyota/hotpot/pkg/storage/ent/do"
	"github.com/dannyota/hotpot/pkg/storage/ent/do/bronzehistorydokubernetescluster"
	"github.com/dannyota/hotpot/pkg/storage/ent/do/bronzehistorydokubernetesnodepool"
)

// ClusterHistoryService handles history tracking for Kubernetes clusters.
type ClusterHistoryService struct {
	entClient *entdo.Client
}

func NewClusterHistoryService(entClient *entdo.Client) *ClusterHistoryService {
	return &ClusterHistoryService{entClient: entClient}
}

func (h *ClusterHistoryService) buildCreate(tx *entdo.Tx, data *ClusterData) *entdo.BronzeHistoryDOKubernetesClusterCreate {
	return tx.BronzeHistoryDOKubernetesCluster.Create().
		SetResourceID(data.ResourceID).
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
		SetNillableAPIUpdatedAt(data.APIUpdatedAt)
}

func (h *ClusterHistoryService) CreateHistory(ctx context.Context, tx *entdo.Tx, data *ClusterData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create KubernetesCluster history: %w", err)
	}
	return nil
}

func (h *ClusterHistoryService) UpdateHistory(ctx context.Context, tx *entdo.Tx, old *entdo.BronzeDOKubernetesCluster, new *ClusterData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKubernetesCluster.Query().
		Where(
			bronzehistorydokubernetescluster.ResourceID(old.ID),
			bronzehistorydokubernetescluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current KubernetesCluster history: %w", err)
	}

	if err := tx.BronzeHistoryDOKubernetesCluster.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close KubernetesCluster history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new KubernetesCluster history: %w", err)
	}

	return nil
}

func (h *ClusterHistoryService) CloseHistory(ctx context.Context, tx *entdo.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKubernetesCluster.Query().
		Where(
			bronzehistorydokubernetescluster.ResourceID(resourceID),
			bronzehistorydokubernetescluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entdo.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current KubernetesCluster history: %w", err)
	}

	if err := tx.BronzeHistoryDOKubernetesCluster.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close KubernetesCluster history: %w", err)
	}

	return nil
}

// NodePoolHistoryService handles history tracking for Kubernetes node pools.
type NodePoolHistoryService struct {
	entClient *entdo.Client
}

func NewNodePoolHistoryService(entClient *entdo.Client) *NodePoolHistoryService {
	return &NodePoolHistoryService{entClient: entClient}
}

func (h *NodePoolHistoryService) buildCreate(tx *entdo.Tx, data *NodePoolData) *entdo.BronzeHistoryDOKubernetesNodePoolCreate {
	return tx.BronzeHistoryDOKubernetesNodePool.Create().
		SetResourceID(data.ResourceID).
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
		SetNodesJSON(data.NodesJSON)
}

func (h *NodePoolHistoryService) CreateHistory(ctx context.Context, tx *entdo.Tx, data *NodePoolData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create KubernetesNodePool history: %w", err)
	}
	return nil
}

func (h *NodePoolHistoryService) UpdateHistory(ctx context.Context, tx *entdo.Tx, old *entdo.BronzeDOKubernetesNodePool, new *NodePoolData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKubernetesNodePool.Query().
		Where(
			bronzehistorydokubernetesnodepool.ResourceID(old.ID),
			bronzehistorydokubernetesnodepool.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current KubernetesNodePool history: %w", err)
	}

	if err := tx.BronzeHistoryDOKubernetesNodePool.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close KubernetesNodePool history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new KubernetesNodePool history: %w", err)
	}

	return nil
}

func (h *NodePoolHistoryService) CloseHistory(ctx context.Context, tx *entdo.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOKubernetesNodePool.Query().
		Where(
			bronzehistorydokubernetesnodepool.ResourceID(resourceID),
			bronzehistorydokubernetesnodepool.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entdo.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current KubernetesNodePool history: %w", err)
	}

	if err := tx.BronzeHistoryDOKubernetesNodePool.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close KubernetesNodePool history: %w", err)
	}

	return nil
}
