package cluster

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcontainercluster"
	"hotpot/pkg/storage/ent/bronzehistorygcpcontainerclusteraddon"
	"hotpot/pkg/storage/ent/bronzehistorygcpcontainerclustercondition"
	"hotpot/pkg/storage/ent/bronzehistorygcpcontainerclusterlabel"
	"hotpot/pkg/storage/ent/bronzehistorygcpcontainerclusternodepool"
)

// HistoryService handles history tracking for clusters.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates history records for a new cluster and all children.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, clusterData *ClusterData, now time.Time) error {
	// Create cluster history
	clusterHistCreate := tx.BronzeHistoryGCPContainerCluster.Create().
		SetResourceID(clusterData.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(clusterData.CollectedAt).
		SetName(clusterData.Name).
		SetLocation(clusterData.Location).
		SetZone(clusterData.Zone).
		SetDescription(clusterData.Description).
		SetInitialClusterVersion(clusterData.InitialClusterVersion).
		SetCurrentMasterVersion(clusterData.CurrentMasterVersion).
		SetCurrentNodeVersion(clusterData.CurrentNodeVersion).
		SetStatus(clusterData.Status).
		SetStatusMessage(clusterData.StatusMessage).
		SetCurrentNodeCount(clusterData.CurrentNodeCount).
		SetNetwork(clusterData.Network).
		SetSubnetwork(clusterData.Subnetwork).
		SetClusterIpv4Cidr(clusterData.ClusterIpv4Cidr).
		SetServicesIpv4Cidr(clusterData.ServicesIpv4Cidr).
		SetNodeIpv4CidrSize(clusterData.NodeIpv4CidrSize).
		SetEndpoint(clusterData.Endpoint).
		SetSelfLink(clusterData.SelfLink).
		SetCreateTime(clusterData.CreateTime).
		SetExpireTime(clusterData.ExpireTime).
		SetEtag(clusterData.Etag).
		SetLabelFingerprint(clusterData.LabelFingerprint).
		SetLoggingService(clusterData.LoggingService).
		SetMonitoringService(clusterData.MonitoringService).
		SetEnableKubernetesAlpha(clusterData.EnableKubernetesAlpha).
		SetEnableTpu(clusterData.EnableTpu).
		SetTpuIpv4CidrBlock(clusterData.TpuIpv4CidrBlock).
		SetProjectID(clusterData.ProjectID)

	// Set optional JSON fields
	if clusterData.AddonsConfigJSON != nil {
		clusterHistCreate.SetAddonsConfigJSON(clusterData.AddonsConfigJSON)
	}
	if clusterData.PrivateClusterConfigJSON != nil {
		clusterHistCreate.SetPrivateClusterConfigJSON(clusterData.PrivateClusterConfigJSON)
	}
	if clusterData.IPAllocationPolicyJSON != nil {
		clusterHistCreate.SetIPAllocationPolicyJSON(clusterData.IPAllocationPolicyJSON)
	}
	if clusterData.NetworkConfigJSON != nil {
		clusterHistCreate.SetNetworkConfigJSON(clusterData.NetworkConfigJSON)
	}
	if clusterData.MasterAuthJSON != nil {
		clusterHistCreate.SetMasterAuthJSON(clusterData.MasterAuthJSON)
	}
	if clusterData.AutoscalingJSON != nil {
		clusterHistCreate.SetAutoscalingJSON(clusterData.AutoscalingJSON)
	}
	if clusterData.VerticalPodAutoscalingJSON != nil {
		clusterHistCreate.SetVerticalPodAutoscalingJSON(clusterData.VerticalPodAutoscalingJSON)
	}
	if clusterData.MonitoringConfigJSON != nil {
		clusterHistCreate.SetMonitoringConfigJSON(clusterData.MonitoringConfigJSON)
	}
	if clusterData.LoggingConfigJSON != nil {
		clusterHistCreate.SetLoggingConfigJSON(clusterData.LoggingConfigJSON)
	}
	if clusterData.MaintenancePolicyJSON != nil {
		clusterHistCreate.SetMaintenancePolicyJSON(clusterData.MaintenancePolicyJSON)
	}
	if clusterData.DatabaseEncryptionJSON != nil {
		clusterHistCreate.SetDatabaseEncryptionJSON(clusterData.DatabaseEncryptionJSON)
	}
	if clusterData.WorkloadIdentityConfigJSON != nil {
		clusterHistCreate.SetWorkloadIdentityConfigJSON(clusterData.WorkloadIdentityConfigJSON)
	}
	if clusterData.AutopilotJSON != nil {
		clusterHistCreate.SetAutopilotJSON(clusterData.AutopilotJSON)
	}
	if clusterData.ReleaseChannelJSON != nil {
		clusterHistCreate.SetReleaseChannelJSON(clusterData.ReleaseChannelJSON)
	}
	if clusterData.BinaryAuthorizationJSON != nil {
		clusterHistCreate.SetBinaryAuthorizationJSON(clusterData.BinaryAuthorizationJSON)
	}
	if clusterData.SecurityPostureConfigJSON != nil {
		clusterHistCreate.SetSecurityPostureConfigJSON(clusterData.SecurityPostureConfigJSON)
	}
	if clusterData.NodePoolDefaultsJSON != nil {
		clusterHistCreate.SetNodePoolDefaultsJSON(clusterData.NodePoolDefaultsJSON)
	}
	if clusterData.FleetJSON != nil {
		clusterHistCreate.SetFleetJSON(clusterData.FleetJSON)
	}

	clusterHist, err := clusterHistCreate.Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create cluster history: %w", err)
	}

	// Create children history with cluster_history_id
	return h.createChildrenHistory(ctx, tx, clusterHist.HistoryID, clusterData, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPContainerCluster, new *ClusterData, diff *ClusterDiff, now time.Time) error {
	// Get current cluster history
	currentHist, err := tx.BronzeHistoryGCPContainerCluster.Query().
		Where(
			bronzehistorygcpcontainercluster.ResourceID(old.ID),
			bronzehistorygcpcontainercluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current cluster history: %w", err)
	}

	// If cluster-level fields changed, close old and create new cluster history
	if diff.IsChanged {
		// Close old cluster history
		if err := tx.BronzeHistoryGCPContainerCluster.UpdateOne(currentHist).
			SetValidTo(now).
			Exec(ctx); err != nil {
			return fmt.Errorf("failed to close cluster history: %w", err)
		}

		// Create new cluster history
		clusterHistCreate := tx.BronzeHistoryGCPContainerCluster.Create().
			SetResourceID(new.ResourceID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetName(new.Name).
			SetLocation(new.Location).
			SetZone(new.Zone).
			SetDescription(new.Description).
			SetInitialClusterVersion(new.InitialClusterVersion).
			SetCurrentMasterVersion(new.CurrentMasterVersion).
			SetCurrentNodeVersion(new.CurrentNodeVersion).
			SetStatus(new.Status).
			SetStatusMessage(new.StatusMessage).
			SetCurrentNodeCount(new.CurrentNodeCount).
			SetNetwork(new.Network).
			SetSubnetwork(new.Subnetwork).
			SetClusterIpv4Cidr(new.ClusterIpv4Cidr).
			SetServicesIpv4Cidr(new.ServicesIpv4Cidr).
			SetNodeIpv4CidrSize(new.NodeIpv4CidrSize).
			SetEndpoint(new.Endpoint).
			SetSelfLink(new.SelfLink).
			SetCreateTime(new.CreateTime).
			SetExpireTime(new.ExpireTime).
			SetEtag(new.Etag).
			SetLabelFingerprint(new.LabelFingerprint).
			SetLoggingService(new.LoggingService).
			SetMonitoringService(new.MonitoringService).
			SetEnableKubernetesAlpha(new.EnableKubernetesAlpha).
			SetEnableTpu(new.EnableTpu).
			SetTpuIpv4CidrBlock(new.TpuIpv4CidrBlock).
			SetProjectID(new.ProjectID)

		// Set optional JSON fields
		if new.AddonsConfigJSON != nil {
			clusterHistCreate.SetAddonsConfigJSON(new.AddonsConfigJSON)
		}
		if new.PrivateClusterConfigJSON != nil {
			clusterHistCreate.SetPrivateClusterConfigJSON(new.PrivateClusterConfigJSON)
		}
		if new.IPAllocationPolicyJSON != nil {
			clusterHistCreate.SetIPAllocationPolicyJSON(new.IPAllocationPolicyJSON)
		}
		if new.NetworkConfigJSON != nil {
			clusterHistCreate.SetNetworkConfigJSON(new.NetworkConfigJSON)
		}
		if new.MasterAuthJSON != nil {
			clusterHistCreate.SetMasterAuthJSON(new.MasterAuthJSON)
		}
		if new.AutoscalingJSON != nil {
			clusterHistCreate.SetAutoscalingJSON(new.AutoscalingJSON)
		}
		if new.VerticalPodAutoscalingJSON != nil {
			clusterHistCreate.SetVerticalPodAutoscalingJSON(new.VerticalPodAutoscalingJSON)
		}
		if new.MonitoringConfigJSON != nil {
			clusterHistCreate.SetMonitoringConfigJSON(new.MonitoringConfigJSON)
		}
		if new.LoggingConfigJSON != nil {
			clusterHistCreate.SetLoggingConfigJSON(new.LoggingConfigJSON)
		}
		if new.MaintenancePolicyJSON != nil {
			clusterHistCreate.SetMaintenancePolicyJSON(new.MaintenancePolicyJSON)
		}
		if new.DatabaseEncryptionJSON != nil {
			clusterHistCreate.SetDatabaseEncryptionJSON(new.DatabaseEncryptionJSON)
		}
		if new.WorkloadIdentityConfigJSON != nil {
			clusterHistCreate.SetWorkloadIdentityConfigJSON(new.WorkloadIdentityConfigJSON)
		}
		if new.AutopilotJSON != nil {
			clusterHistCreate.SetAutopilotJSON(new.AutopilotJSON)
		}
		if new.ReleaseChannelJSON != nil {
			clusterHistCreate.SetReleaseChannelJSON(new.ReleaseChannelJSON)
		}
		if new.BinaryAuthorizationJSON != nil {
			clusterHistCreate.SetBinaryAuthorizationJSON(new.BinaryAuthorizationJSON)
		}
		if new.SecurityPostureConfigJSON != nil {
			clusterHistCreate.SetSecurityPostureConfigJSON(new.SecurityPostureConfigJSON)
		}
		if new.NodePoolDefaultsJSON != nil {
			clusterHistCreate.SetNodePoolDefaultsJSON(new.NodePoolDefaultsJSON)
		}
		if new.FleetJSON != nil {
			clusterHistCreate.SetFleetJSON(new.FleetJSON)
		}

		clusterHist, err := clusterHistCreate.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new cluster history: %w", err)
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now); err != nil {
			return fmt.Errorf("failed to close children history: %w", err)
		}
		return h.createChildrenHistory(ctx, tx, clusterHist.HistoryID, new, now)
	}

	// Cluster unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(ctx, tx, currentHist.HistoryID, old, new, diff, now)
}

// CloseHistory closes history records for a deleted cluster.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current cluster history
	currentHist, err := tx.BronzeHistoryGCPContainerCluster.Query().
		Where(
			bronzehistorygcpcontainercluster.ResourceID(resourceID),
			bronzehistorygcpcontainercluster.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current cluster history: %w", err)
	}

	// Close cluster history
	if err := tx.BronzeHistoryGCPContainerCluster.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("failed to close cluster history: %w", err)
	}

	// Close all children history
	return h.closeChildrenHistory(ctx, tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, data *ClusterData, now time.Time) error {
	// Labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeHistoryGCPContainerClusterLabel.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	// Addons
	for _, addonData := range data.Addons {
		create := tx.BronzeHistoryGCPContainerClusterAddon.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetAddonName(addonData.AddonName).
			SetEnabled(addonData.Enabled)

		if addonData.ConfigJSON != nil {
			create.SetConfigJSON(addonData.ConfigJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create addon history: %w", err)
		}
	}

	// Conditions
	for _, condData := range data.Conditions {
		_, err := tx.BronzeHistoryGCPContainerClusterCondition.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetCode(condData.Code).
			SetMessage(condData.Message).
			SetCanonicalCode(condData.CanonicalCode).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create condition history: %w", err)
		}
	}

	// Node pools
	for _, npData := range data.NodePools {
		create := tx.BronzeHistoryGCPContainerClusterNodePool.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetName(npData.Name).
			SetVersion(npData.Version).
			SetStatus(npData.Status).
			SetStatusMessage(npData.StatusMessage).
			SetInitialNodeCount(npData.InitialNodeCount).
			SetSelfLink(npData.SelfLink).
			SetPodIpv4CidrSize(npData.PodIpv4CidrSize).
			SetEtag(npData.Etag)

		if npData.LocationsJSON != nil {
			create.SetLocationsJSON(npData.LocationsJSON)
		}
		if npData.ConfigJSON != nil {
			create.SetConfigJSON(npData.ConfigJSON)
		}
		if npData.AutoscalingJSON != nil {
			create.SetAutoscalingJSON(npData.AutoscalingJSON)
		}
		if npData.ManagementJSON != nil {
			create.SetManagementJSON(npData.ManagementJSON)
		}
		if npData.UpgradeSettingsJSON != nil {
			create.SetUpgradeSettingsJSON(npData.UpgradeSettingsJSON)
		}
		if npData.NetworkConfigJSON != nil {
			create.SetNetworkConfigJSON(npData.NetworkConfigJSON)
		}
		if npData.PlacementPolicyJSON != nil {
			create.SetPlacementPolicyJSON(npData.PlacementPolicyJSON)
		}
		if npData.MaxPodsConstraintJSON != nil {
			create.SetMaxPodsConstraintJSON(npData.MaxPodsConstraintJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create node pool history: %w", err)
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, now time.Time) error {
	// Labels
	_, err := tx.BronzeHistoryGCPContainerClusterLabel.Update().
		Where(
			bronzehistorygcpcontainerclusterlabel.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusterlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Addons
	_, err = tx.BronzeHistoryGCPContainerClusterAddon.Update().
		Where(
			bronzehistorygcpcontainerclusteraddon.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusteraddon.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close addon history: %w", err)
	}

	// Conditions
	_, err = tx.BronzeHistoryGCPContainerClusterCondition.Update().
		Where(
			bronzehistorygcpcontainerclustercondition.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclustercondition.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close condition history: %w", err)
	}

	// Node pools
	_, err = tx.BronzeHistoryGCPContainerClusterNodePool.Update().
		Where(
			bronzehistorygcpcontainerclusternodepool.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusternodepool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close node pool history: %w", err)
	}

	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, old *ent.BronzeGCPContainerCluster, new *ClusterData, diff *ClusterDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(ctx, tx, clusterHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.AddonsDiff.Changed {
		if err := h.updateAddonsHistory(ctx, tx, clusterHistoryID, new.Addons, now); err != nil {
			return err
		}
	}

	if diff.ConditionsDiff.Changed {
		if err := h.updateConditionsHistory(ctx, tx, clusterHistoryID, new.Conditions, now); err != nil {
			return err
		}
	}

	if diff.NodePoolsDiff.Changed {
		if err := h.updateNodePoolsHistory(ctx, tx, clusterHistoryID, new.NodePools, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, labels []LabelData, now time.Time) error {
	// Close old label history
	_, err := tx.BronzeHistoryGCPContainerClusterLabel.Update().
		Where(
			bronzehistorygcpcontainerclusterlabel.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusterlabel.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close label history: %w", err)
	}

	// Create new label history
	for _, labelData := range labels {
		_, err := tx.BronzeHistoryGCPContainerClusterLabel.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateAddonsHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, addons []AddonData, now time.Time) error {
	// Close old addon history
	_, err := tx.BronzeHistoryGCPContainerClusterAddon.Update().
		Where(
			bronzehistorygcpcontainerclusteraddon.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusteraddon.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close addon history: %w", err)
	}

	// Create new addon history
	for _, addonData := range addons {
		create := tx.BronzeHistoryGCPContainerClusterAddon.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetAddonName(addonData.AddonName).
			SetEnabled(addonData.Enabled)

		if addonData.ConfigJSON != nil {
			create.SetConfigJSON(addonData.ConfigJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create addon history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateConditionsHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, conditions []ConditionData, now time.Time) error {
	// Close old condition history
	_, err := tx.BronzeHistoryGCPContainerClusterCondition.Update().
		Where(
			bronzehistorygcpcontainerclustercondition.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclustercondition.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close condition history: %w", err)
	}

	// Create new condition history
	for _, condData := range conditions {
		_, err := tx.BronzeHistoryGCPContainerClusterCondition.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetCode(condData.Code).
			SetMessage(condData.Message).
			SetCanonicalCode(condData.CanonicalCode).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create condition history: %w", err)
		}
	}

	return nil
}

func (h *HistoryService) updateNodePoolsHistory(ctx context.Context, tx *ent.Tx, clusterHistoryID uint, nodePools []NodePoolData, now time.Time) error {
	// Close old node pool history
	_, err := tx.BronzeHistoryGCPContainerClusterNodePool.Update().
		Where(
			bronzehistorygcpcontainerclusternodepool.ClusterHistoryID(clusterHistoryID),
			bronzehistorygcpcontainerclusternodepool.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close node pool history: %w", err)
	}

	// Create new node pool history
	for _, npData := range nodePools {
		create := tx.BronzeHistoryGCPContainerClusterNodePool.Create().
			SetClusterHistoryID(clusterHistoryID).
			SetValidFrom(now).
			SetName(npData.Name).
			SetVersion(npData.Version).
			SetStatus(npData.Status).
			SetStatusMessage(npData.StatusMessage).
			SetInitialNodeCount(npData.InitialNodeCount).
			SetSelfLink(npData.SelfLink).
			SetPodIpv4CidrSize(npData.PodIpv4CidrSize).
			SetEtag(npData.Etag)

		if npData.LocationsJSON != nil {
			create.SetLocationsJSON(npData.LocationsJSON)
		}
		if npData.ConfigJSON != nil {
			create.SetConfigJSON(npData.ConfigJSON)
		}
		if npData.AutoscalingJSON != nil {
			create.SetAutoscalingJSON(npData.AutoscalingJSON)
		}
		if npData.ManagementJSON != nil {
			create.SetManagementJSON(npData.ManagementJSON)
		}
		if npData.UpgradeSettingsJSON != nil {
			create.SetUpgradeSettingsJSON(npData.UpgradeSettingsJSON)
		}
		if npData.NetworkConfigJSON != nil {
			create.SetNetworkConfigJSON(npData.NetworkConfigJSON)
		}
		if npData.PlacementPolicyJSON != nil {
			create.SetPlacementPolicyJSON(npData.PlacementPolicyJSON)
		}
		if npData.MaxPodsConstraintJSON != nil {
			create.SetMaxPodsConstraintJSON(npData.MaxPodsConstraintJSON)
		}

		_, err := create.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create node pool history: %w", err)
		}
	}

	return nil
}
