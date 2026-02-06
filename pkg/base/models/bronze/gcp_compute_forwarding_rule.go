package bronze

import (
	"time"

	"hotpot/pkg/base/jsonb"
)

// GCPComputeForwardingRule represents a GCP Compute Engine regional forwarding rule in the bronze layer.
// Fields preserve raw API response data from compute.forwardingRules.aggregatedList.
type GCPComputeForwardingRule struct {
	// GCP API fields (json tag = original API field name for traceability)
	// ResourceID is the GCP API ID, used as primary key for linking.
	ResourceID                                          string  `gorm:"primaryKey;column:resource_id;type:varchar(255)" json:"id"`
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

	// JSONB fields (primitive arrays and nested objects)
	PortsJSON                          jsonb.JSON `gorm:"column:ports_json;type:jsonb" json:"ports"`
	SourceIpRangesJSON                 jsonb.JSON `gorm:"column:source_ip_ranges_json;type:jsonb" json:"sourceIpRanges"`
	MetadataFiltersJSON                jsonb.JSON `gorm:"column:metadata_filters_json;type:jsonb" json:"metadataFilters"`
	ServiceDirectoryRegistrationsJSON  jsonb.JSON `gorm:"column:service_directory_registrations_json;type:jsonb" json:"serviceDirectoryRegistrations"`

	// Collection metadata
	ProjectID   string    `gorm:"column:project_id;type:varchar(255);not null;index" json:"-"`
	CollectedAt time.Time `gorm:"column:collected_at;not null;index" json:"-"`

	// Relationships (linked by ResourceID, no FK constraint)
	Labels []GCPComputeForwardingRuleLabel `gorm:"foreignKey:ForwardingRuleResourceID;references:ResourceID" json:"labels,omitempty"`
}

func (GCPComputeForwardingRule) TableName() string {
	return "bronze.gcp_compute_forwarding_rules"
}

// GCPComputeForwardingRuleLabel represents a label attached to a GCP Compute forwarding rule.
type GCPComputeForwardingRuleLabel struct {
	ID                         uint   `gorm:"primaryKey"`
	ForwardingRuleResourceID   string `gorm:"column:forwarding_rule_resource_id;type:varchar(255);not null;index" json:"-"`
	Key                        string `gorm:"column:key;type:varchar(255);not null" json:"key"`
	Value                      string `gorm:"column:value;type:text" json:"value"`
}

func (GCPComputeForwardingRuleLabel) TableName() string {
	return "bronze.gcp_compute_forwarding_rule_labels"
}
