package bronze_history

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeGlobalForwardingRule stores historical snapshots of GCP Compute global forwarding rules.
// Uses ResourceID for lookup (has API ID), with valid_from/valid_to for time range.
type GCPComputeGlobalForwardingRule struct {
	HistoryID uint `gorm:"primaryKey"`

	// Link by API ID (ForwardingRule has unique API ID)
	ResourceID string `gorm:"column:resource_id;type:varchar(255);not null;index" json:"id"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"` // null = current

	// All forwarding rule fields (same as bronze.GCPComputeGlobalForwardingRule)
	Name                                                string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Description                                         string  `gorm:"column:description;type:text" json:"description"`
	IPAddress                                           string  `gorm:"column:ip_address;type:varchar(50)" json:"IPAddress"`
	IPProtocol                                          string  `gorm:"column:ip_protocol;type:varchar(20)" json:"IPProtocol"`
	AllPorts                                            bool    `gorm:"column:all_ports" json:"allPorts"`
	AllowGlobalAccess                                   bool    `gorm:"column:allow_global_access" json:"allowGlobalAccess"`
	AllowPscGlobalAccess                                bool    `gorm:"column:allow_psc_global_access" json:"allowPscGlobalAccess"`
	BackendService                                      string  `gorm:"column:backend_service;type:text" json:"backendService"`
	BaseForwardingRule                                  string  `gorm:"column:base_forwarding_rule;type:text" json:"baseForwardingRule"`
	CreationTimestamp                                   string  `gorm:"column:creation_timestamp;type:varchar(50)" json:"creationTimestamp"`
	ExternalManagedBackendBucketMigrationState          string  `gorm:"column:external_managed_backend_bucket_migration_state;type:varchar(50)" json:"externalManagedBackendBucketMigrationState"`
	ExternalManagedBackendBucketMigrationTestingPercentage float32 `gorm:"column:external_managed_backend_bucket_migration_testing_percentage;type:real" json:"externalManagedBackendBucketMigrationTestingPercentage"`
	Fingerprint                                         string  `gorm:"column:fingerprint;type:varchar(255)" json:"fingerprint"`
	IpCollection                                        string  `gorm:"column:ip_collection;type:text" json:"ipCollection"`
	IpVersion                                           string  `gorm:"column:ip_version;type:varchar(10)" json:"ipVersion"`
	IsMirroringCollector                                bool    `gorm:"column:is_mirroring_collector" json:"isMirroringCollector"`
	LabelFingerprint                                    string  `gorm:"column:label_fingerprint;type:varchar(255)" json:"labelFingerprint"`
	LoadBalancingScheme                                 string  `gorm:"column:load_balancing_scheme;type:varchar(50)" json:"loadBalancingScheme"`
	Network                                             string  `gorm:"column:network;type:text" json:"network"`
	NetworkTier                                         string  `gorm:"column:network_tier;type:varchar(50)" json:"networkTier"`
	NoAutomateDnsZone                                   bool    `gorm:"column:no_automate_dns_zone" json:"noAutomateDnsZone"`
	PortRange                                           string  `gorm:"column:port_range;type:varchar(50)" json:"portRange"`
	PscConnectionId                                     string  `gorm:"column:psc_connection_id;type:varchar(255)" json:"pscConnectionId"`
	PscConnectionStatus                                 string  `gorm:"column:psc_connection_status;type:varchar(50)" json:"pscConnectionStatus"`
	Region                                              string  `gorm:"column:region;type:text" json:"region"`
	SelfLink                                            string  `gorm:"column:self_link;type:text" json:"selfLink"`
	SelfLinkWithId                                      string  `gorm:"column:self_link_with_id;type:text" json:"selfLinkWithId"`
	ServiceLabel                                        string  `gorm:"column:service_label;type:varchar(255)" json:"serviceLabel"`
	ServiceName                                         string  `gorm:"column:service_name;type:text" json:"serviceName"`
	Subnetwork                                          string  `gorm:"column:subnetwork;type:text" json:"subnetwork"`
	Target                                              string  `gorm:"column:target;type:text" json:"target"`

	// JSONB fields
	PortsJSON                          jsonb.JSON `gorm:"column:ports_json;type:jsonb" json:"ports"`
	SourceIpRangesJSON                 jsonb.JSON `gorm:"column:source_ip_ranges_json;type:jsonb" json:"sourceIpRanges"`
	MetadataFiltersJSON                jsonb.JSON `gorm:"column:metadata_filters_json;type:jsonb" json:"metadataFilters"`
	ServiceDirectoryRegistrationsJSON  jsonb.JSON `gorm:"column:service_directory_registrations_json;type:jsonb" json:"serviceDirectoryRegistrations"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"projectId"`
	CollectedAt time.Time `gorm:"column:collected_at;not null" json:"collectedAt"`
}

func (GCPComputeGlobalForwardingRule) TableName() string {
	return "bronze_history.gcp_compute_global_forwarding_rules"
}

// GCPComputeGlobalForwardingRuleLabel stores historical snapshots of global forwarding rule labels.
// Links via GlobalForwardingRuleHistoryID, has own valid_from/valid_to for granular tracking.
type GCPComputeGlobalForwardingRuleLabel struct {
	HistoryID                      uint `gorm:"primaryKey"`
	GlobalForwardingRuleHistoryID  uint `gorm:"column:global_forwarding_rule_history_id;not null;index" json:"-"`

	// Time range for this version
	ValidFrom time.Time  `gorm:"column:valid_from;not null;index" json:"validFrom"`
	ValidTo   *time.Time `gorm:"column:valid_to;index" json:"validTo"`

	// Label fields
	Key   string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeGlobalForwardingRuleLabel) TableName() string {
	return "bronze_history.gcp_compute_global_forwarding_rule_labels"
}
