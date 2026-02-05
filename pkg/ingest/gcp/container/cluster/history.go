package cluster

import (
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
	"hotpot/pkg/base/models/bronze_history"
)

// HistoryService handles history tracking for clusters.
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service.
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// CreateHistory creates history records for a new cluster and all children.
func (h *HistoryService) CreateHistory(tx *gorm.DB, cluster *bronze.GCPContainerCluster, now time.Time) error {
	// Create cluster history
	clusterHist := toClusterHistory(cluster, now)
	if err := tx.Create(&clusterHist).Error; err != nil {
		return err
	}

	// Create children history with cluster_history_id
	return h.createChildrenHistory(tx, clusterHist.HistoryID, cluster, now)
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(tx *gorm.DB, old, new *bronze.GCPContainerCluster, diff *ClusterDiff, now time.Time) error {
	// Get current cluster history
	var currentHist bronze_history.GCPContainerCluster
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", old.ResourceID).First(&currentHist).Error; err != nil {
		return err
	}

	// If cluster-level fields changed, close old and create new cluster history
	if diff.IsChanged {
		// Close old cluster history
		if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
			return err
		}

		// Create new cluster history
		clusterHist := toClusterHistory(new, now)
		if err := tx.Create(&clusterHist).Error; err != nil {
			return err
		}

		// Close all children history and create new ones
		if err := h.closeChildrenHistory(tx, currentHist.HistoryID, now); err != nil {
			return err
		}
		return h.createChildrenHistory(tx, clusterHist.HistoryID, new, now)
	}

	// Cluster unchanged, check children individually (granular tracking)
	return h.updateChildrenHistory(tx, currentHist.HistoryID, new, diff, now)
}

// CloseHistory closes history records for a deleted cluster.
func (h *HistoryService) CloseHistory(tx *gorm.DB, resourceID string, now time.Time) error {
	// Get current cluster history
	var currentHist bronze_history.GCPContainerCluster
	if err := tx.Where("resource_id = ? AND valid_to IS NULL", resourceID).First(&currentHist).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil // No history to close
		}
		return err
	}

	// Close cluster history
	if err := tx.Model(&currentHist).Update("valid_to", now).Error; err != nil {
		return err
	}

	// Close all children history
	return h.closeChildrenHistory(tx, currentHist.HistoryID, now)
}

// createChildrenHistory creates history records for all children.
func (h *HistoryService) createChildrenHistory(tx *gorm.DB, clusterHistoryID uint, cluster *bronze.GCPContainerCluster, now time.Time) error {
	// Labels
	for _, label := range cluster.Labels {
		labelHist := toLabelHistory(&label, clusterHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}

	// Addons
	for _, addon := range cluster.Addons {
		addonHist := toAddonHistory(&addon, clusterHistoryID, now)
		if err := tx.Create(&addonHist).Error; err != nil {
			return err
		}
	}

	// Conditions
	for _, cond := range cluster.Conditions {
		condHist := toConditionHistory(&cond, clusterHistoryID, now)
		if err := tx.Create(&condHist).Error; err != nil {
			return err
		}
	}

	// Node pools
	for _, np := range cluster.NodePools {
		npHist := toNodePoolHistory(&np, clusterHistoryID, now)
		if err := tx.Create(&npHist).Error; err != nil {
			return err
		}
	}

	return nil
}

// closeChildrenHistory closes all children history records.
func (h *HistoryService) closeChildrenHistory(tx *gorm.DB, clusterHistoryID uint, now time.Time) error {
	tables := []string{
		"bronze_history.gcp_container_cluster_labels",
		"bronze_history.gcp_container_cluster_addons",
		"bronze_history.gcp_container_cluster_conditions",
		"bronze_history.gcp_container_cluster_node_pools",
	}
	for _, table := range tables {
		if err := tx.Table(table).
			Where("cluster_history_id = ? AND valid_to IS NULL", clusterHistoryID).
			Update("valid_to", now).Error; err != nil {
			return err
		}
	}
	return nil
}

// updateChildrenHistory updates children history based on diff (granular tracking).
func (h *HistoryService) updateChildrenHistory(tx *gorm.DB, clusterHistoryID uint, new *bronze.GCPContainerCluster, diff *ClusterDiff, now time.Time) error {
	if diff.LabelsDiff.Changed {
		if err := h.updateLabelsHistory(tx, clusterHistoryID, new.Labels, now); err != nil {
			return err
		}
	}

	if diff.AddonsDiff.Changed {
		if err := h.updateAddonsHistory(tx, clusterHistoryID, new.Addons, now); err != nil {
			return err
		}
	}

	if diff.ConditionsDiff.Changed {
		if err := h.updateConditionsHistory(tx, clusterHistoryID, new.Conditions, now); err != nil {
			return err
		}
	}

	if diff.NodePoolsDiff.Changed {
		if err := h.updateNodePoolsHistory(tx, clusterHistoryID, new.NodePools, now); err != nil {
			return err
		}
	}

	return nil
}

func (h *HistoryService) updateLabelsHistory(tx *gorm.DB, clusterHistoryID uint, labels []bronze.GCPContainerClusterLabel, now time.Time) error {
	tx.Table("bronze_history.gcp_container_cluster_labels").
		Where("cluster_history_id = ? AND valid_to IS NULL", clusterHistoryID).
		Update("valid_to", now)

	for _, label := range labels {
		labelHist := toLabelHistory(&label, clusterHistoryID, now)
		if err := tx.Create(&labelHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateAddonsHistory(tx *gorm.DB, clusterHistoryID uint, addons []bronze.GCPContainerClusterAddon, now time.Time) error {
	tx.Table("bronze_history.gcp_container_cluster_addons").
		Where("cluster_history_id = ? AND valid_to IS NULL", clusterHistoryID).
		Update("valid_to", now)

	for _, addon := range addons {
		addonHist := toAddonHistory(&addon, clusterHistoryID, now)
		if err := tx.Create(&addonHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateConditionsHistory(tx *gorm.DB, clusterHistoryID uint, conditions []bronze.GCPContainerClusterCondition, now time.Time) error {
	tx.Table("bronze_history.gcp_container_cluster_conditions").
		Where("cluster_history_id = ? AND valid_to IS NULL", clusterHistoryID).
		Update("valid_to", now)

	for _, cond := range conditions {
		condHist := toConditionHistory(&cond, clusterHistoryID, now)
		if err := tx.Create(&condHist).Error; err != nil {
			return err
		}
	}
	return nil
}

func (h *HistoryService) updateNodePoolsHistory(tx *gorm.DB, clusterHistoryID uint, nodePools []bronze.GCPContainerClusterNodePool, now time.Time) error {
	tx.Table("bronze_history.gcp_container_cluster_node_pools").
		Where("cluster_history_id = ? AND valid_to IS NULL", clusterHistoryID).
		Update("valid_to", now)

	for _, np := range nodePools {
		npHist := toNodePoolHistory(&np, clusterHistoryID, now)
		if err := tx.Create(&npHist).Error; err != nil {
			return err
		}
	}
	return nil
}

// Conversion functions: bronze -> bronze_history

func toClusterHistory(c *bronze.GCPContainerCluster, now time.Time) bronze_history.GCPContainerCluster {
	return bronze_history.GCPContainerCluster{
		ResourceID:                 c.ResourceID,
		ValidFrom:                  now,
		ValidTo:                    nil,
		Name:                       c.Name,
		Location:                   c.Location,
		Zone:                       c.Zone,
		Description:                c.Description,
		InitialClusterVersion:      c.InitialClusterVersion,
		CurrentMasterVersion:       c.CurrentMasterVersion,
		CurrentNodeVersion:         c.CurrentNodeVersion,
		Status:                     c.Status,
		StatusMessage:              c.StatusMessage,
		CurrentNodeCount:           c.CurrentNodeCount,
		Network:                    c.Network,
		Subnetwork:                 c.Subnetwork,
		ClusterIpv4Cidr:            c.ClusterIpv4Cidr,
		ServicesIpv4Cidr:           c.ServicesIpv4Cidr,
		NodeIpv4CidrSize:           c.NodeIpv4CidrSize,
		Endpoint:                   c.Endpoint,
		SelfLink:                   c.SelfLink,
		CreateTime:                 c.CreateTime,
		ExpireTime:                 c.ExpireTime,
		Etag:                       c.Etag,
		LabelFingerprint:           c.LabelFingerprint,
		LoggingService:             c.LoggingService,
		MonitoringService:          c.MonitoringService,
		EnableKubernetesAlpha:      c.EnableKubernetesAlpha,
		EnableTpu:                  c.EnableTpu,
		TpuIpv4CidrBlock:           c.TpuIpv4CidrBlock,
		AddonsConfigJSON:           c.AddonsConfigJSON,
		PrivateClusterConfigJSON:   c.PrivateClusterConfigJSON,
		IpAllocationPolicyJSON:     c.IpAllocationPolicyJSON,
		NetworkConfigJSON:          c.NetworkConfigJSON,
		MasterAuthJSON:             c.MasterAuthJSON,
		AutoscalingJSON:            c.AutoscalingJSON,
		VerticalPodAutoscalingJSON: c.VerticalPodAutoscalingJSON,
		MonitoringConfigJSON:       c.MonitoringConfigJSON,
		LoggingConfigJSON:          c.LoggingConfigJSON,
		MaintenancePolicyJSON:      c.MaintenancePolicyJSON,
		DatabaseEncryptionJSON:     c.DatabaseEncryptionJSON,
		WorkloadIdentityConfigJSON: c.WorkloadIdentityConfigJSON,
		AutopilotJSON:              c.AutopilotJSON,
		ReleaseChannelJSON:         c.ReleaseChannelJSON,
		BinaryAuthorizationJSON:    c.BinaryAuthorizationJSON,
		SecurityPostureConfigJSON:  c.SecurityPostureConfigJSON,
		NodePoolDefaultsJSON:       c.NodePoolDefaultsJSON,
		FleetJSON:                  c.FleetJSON,
		ProjectID:                  c.ProjectID,
		CollectedAt:                c.CollectedAt,
	}
}

func toLabelHistory(l *bronze.GCPContainerClusterLabel, clusterHistoryID uint, now time.Time) bronze_history.GCPContainerClusterLabel {
	return bronze_history.GCPContainerClusterLabel{
		ClusterHistoryID: clusterHistoryID,
		ValidFrom:        now,
		ValidTo:          nil,
		Key:              l.Key,
		Value:            l.Value,
	}
}

func toAddonHistory(a *bronze.GCPContainerClusterAddon, clusterHistoryID uint, now time.Time) bronze_history.GCPContainerClusterAddon {
	return bronze_history.GCPContainerClusterAddon{
		ClusterHistoryID: clusterHistoryID,
		ValidFrom:        now,
		ValidTo:          nil,
		AddonName:        a.AddonName,
		Enabled:          a.Enabled,
		ConfigJSON:       a.ConfigJSON,
	}
}

func toConditionHistory(c *bronze.GCPContainerClusterCondition, clusterHistoryID uint, now time.Time) bronze_history.GCPContainerClusterCondition {
	return bronze_history.GCPContainerClusterCondition{
		ClusterHistoryID: clusterHistoryID,
		ValidFrom:        now,
		ValidTo:          nil,
		Code:             c.Code,
		Message:          c.Message,
		CanonicalCode:    c.CanonicalCode,
	}
}

func toNodePoolHistory(np *bronze.GCPContainerClusterNodePool, clusterHistoryID uint, now time.Time) bronze_history.GCPContainerClusterNodePool {
	return bronze_history.GCPContainerClusterNodePool{
		ClusterHistoryID:      clusterHistoryID,
		ValidFrom:             now,
		ValidTo:               nil,
		Name:                  np.Name,
		Version:               np.Version,
		Status:                np.Status,
		StatusMessage:         np.StatusMessage,
		InitialNodeCount:      np.InitialNodeCount,
		SelfLink:              np.SelfLink,
		PodIpv4CidrSize:       np.PodIpv4CidrSize,
		Etag:                  np.Etag,
		LocationsJSON:         np.LocationsJSON,
		ConfigJSON:            np.ConfigJSON,
		AutoscalingJSON:       np.AutoscalingJSON,
		ManagementJSON:        np.ManagementJSON,
		UpgradeSettingsJSON:   np.UpgradeSettingsJSON,
		NetworkConfigJSON:     np.NetworkConfigJSON,
		PlacementPolicyJSON:   np.PlacementPolicyJSON,
		MaxPodsConstraintJSON: np.MaxPodsConstraintJSON,
	}
}
