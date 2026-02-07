package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeTargetVpnGateway stores historical versions of Classic VPN gateways.
type GCPComputeTargetVpnGateway struct {
	HistoryID uint `gorm:"primaryKey"`

	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Name              string `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description       string `gorm:"column:description;type:text" json:"description"`
	Status            string `gorm:"column:status;type:varchar(50)" json:"status"`
	Region            string `gorm:"column:region;type:text" json:"region"`
	Network           string `gorm:"column:network;type:text" json:"network"`
	SelfLink          string `gorm:"column:self_link;type:text" json:"selfLink"`
	CreationTimestamp string `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	LabelFingerprint  string `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`

	ForwardingRulesJSON jsonb.JSON `gorm:"column:forwarding_rules_json;type:jsonb" json:"forwardingRules"`
	TunnelsJSON         jsonb.JSON `gorm:"column:tunnels_json;type:jsonb" json:"tunnels"`

	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeTargetVpnGateway) TableName() string {
	return "bronze_history.gcp_compute_target_vpn_gateways"
}

// GCPComputeTargetVpnGatewayLabel stores historical versions of Classic VPN gateway labels.
type GCPComputeTargetVpnGatewayLabel struct {
	HistoryID                 uint `gorm:"primaryKey"`
	TargetVpnGatewayHistoryID uint `gorm:"column:target_vpn_gateway_history_id;not null;index" json:"-"`

	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeTargetVpnGatewayLabel) TableName() string {
	return "bronze_history.gcp_compute_target_vpn_gateway_labels"
}
