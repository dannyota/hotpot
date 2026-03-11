package kubernetes

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/digitalocean/godo"
)

// ClusterData holds converted Kubernetes cluster data ready for Ent insertion.
type ClusterData struct {
	ResourceID               string
	Name                     string
	RegionSlug               string
	VersionSlug              string
	ClusterSubnet            string
	ServiceSubnet            string
	IPv4                     string
	Endpoint                 string
	VPCUUID                  string
	HA                       bool
	AutoUpgrade              bool
	SurgeUpgrade             bool
	RegistryEnabled          bool
	StatusState              string
	StatusMessage            string
	TagsJSON                 json.RawMessage
	MaintenancePolicyJSON    json.RawMessage
	ControlPlaneFirewallJSON json.RawMessage
	AutoscalerConfigJSON     json.RawMessage
	APICreatedAt             *time.Time
	APIUpdatedAt             *time.Time
	CollectedAt              time.Time
}

// ConvertCluster converts a godo KubernetesCluster to ClusterData.
func ConvertCluster(v *godo.KubernetesCluster, collectedAt time.Time) *ClusterData {
	data := &ClusterData{
		ResourceID:      v.ID,
		Name:            v.Name,
		RegionSlug:      v.RegionSlug,
		VersionSlug:     v.VersionSlug,
		ClusterSubnet:   v.ClusterSubnet,
		ServiceSubnet:   v.ServiceSubnet,
		IPv4:            v.IPv4,
		Endpoint:        v.Endpoint,
		VPCUUID:         v.VPCUUID,
		HA:              v.HA,
		AutoUpgrade:     v.AutoUpgrade,
		SurgeUpgrade:    v.SurgeUpgrade,
		RegistryEnabled: v.RegistryEnabled,
		CollectedAt:     collectedAt,
	}

	if v.Status != nil {
		data.StatusState = string(v.Status.State)
		data.StatusMessage = v.Status.Message
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if v.MaintenancePolicy != nil {
		data.MaintenancePolicyJSON, _ = json.Marshal(v.MaintenancePolicy)
	}

	if v.ControlPlaneFirewall != nil {
		data.ControlPlaneFirewallJSON, _ = json.Marshal(v.ControlPlaneFirewall)
	}

	if v.ClusterAutoscalerConfiguration != nil {
		data.AutoscalerConfigJSON, _ = json.Marshal(v.ClusterAutoscalerConfiguration)
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	if !v.UpdatedAt.IsZero() {
		t := v.UpdatedAt
		data.APIUpdatedAt = &t
	}

	return data
}

// NodePoolData holds converted Kubernetes node pool data ready for Ent insertion.
type NodePoolData struct {
	ResourceID  string
	ClusterID   string
	NodePoolID  string
	Name        string
	Size        string
	Count       int
	AutoScale   bool
	MinNodes    int
	MaxNodes    int
	TagsJSON    json.RawMessage
	LabelsJSON  json.RawMessage
	TaintsJSON  json.RawMessage
	NodesJSON   json.RawMessage
	CollectedAt time.Time
}

// ConvertNodePool converts a godo KubernetesNodePool to NodePoolData.
func ConvertNodePool(v *godo.KubernetesNodePool, clusterID string, collectedAt time.Time) *NodePoolData {
	data := &NodePoolData{
		ResourceID:  fmt.Sprintf("%s:%s", clusterID, v.ID),
		ClusterID:   clusterID,
		NodePoolID:  v.ID,
		Name:        v.Name,
		Size:        v.Size,
		Count:       v.Count,
		AutoScale:   v.AutoScale,
		MinNodes:    v.MinNodes,
		MaxNodes:    v.MaxNodes,
		CollectedAt: collectedAt,
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if len(v.Labels) > 0 {
		data.LabelsJSON, _ = json.Marshal(v.Labels)
	}

	if len(v.Taints) > 0 {
		data.TaintsJSON, _ = json.Marshal(v.Taints)
	}

	if len(v.Nodes) > 0 {
		data.NodesJSON, _ = json.Marshal(v.Nodes)
	}

	return data
}
