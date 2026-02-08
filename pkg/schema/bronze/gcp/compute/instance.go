package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeInstance represents a GCP Compute Engine instance in the bronze layer.
// Fields preserve raw API response data from compute.instances.list.
type BronzeGCPComputeInstance struct {
	ent.Schema
}

func (BronzeGCPComputeInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeInstance) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields (preserving original API field structure)
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("zone").
			Optional(),
		field.String("machine_type").
			Optional(),
		field.String("status").
			Optional(),
		field.String("status_message").
			Optional(),
		field.String("cpu_platform").
			Optional(),
		field.String("hostname").
			Optional(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("last_start_timestamp").
			Optional(),
		field.String("last_stop_timestamp").
			Optional(),
		field.String("last_suspended_timestamp").
			Optional(),
		field.Bool("deletion_protection").
			Default(false),
		field.Bool("can_ip_forward").
			Default(false),
		field.String("self_link").
			Optional(),

		// SchedulingJSON contains VM scheduling configuration.
		//
		//	{
		//	  "preemptible": bool,
		//	  "onHostMaintenance": "MIGRATE" | "TERMINATE",
		//	  "automaticRestart": bool,
		//	  "provisioningModel": "STANDARD" | "SPOT"
		//	}
		field.JSON("scheduling_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("disks", BronzeGCPComputeInstanceDisk.Type),
		edge.To("nics", BronzeGCPComputeInstanceNIC.Type),
		edge.To("labels", BronzeGCPComputeInstanceLabel.Type),
		edge.To("tags", BronzeGCPComputeInstanceTag.Type),
		edge.To("metadata", BronzeGCPComputeInstanceMetadata.Type),
		edge.To("service_accounts", BronzeGCPComputeInstanceServiceAccount.Type),
	}
}

func (BronzeGCPComputeInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instances"},
	}
}

// BronzeGCPComputeInstanceDisk represents an attached disk on a GCP Compute instance.
// This stores attachment info from instance.disks[], not full disk resource data.
type BronzeGCPComputeInstanceDisk struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceDisk) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields (preserving original API field structure)
		field.String("source").
			Optional(),
		field.String("device_name").
			Optional(),
		field.Int("index").
			Optional(),
		field.Bool("boot").
			Default(false),
		field.Bool("auto_delete").
			Default(false),
		field.String("mode").
			Optional(),
		field.String("interface").
			Optional(),
		field.String("type").
			Optional(),
		field.Int64("disk_size_gb").
			Optional(),

		// DiskEncryptionKeyJSON contains disk encryption configuration.
		//
		//	{
		//	  "kmsKeyName": "projects/.../cryptoKeys/...",
		//	  "sha256": "base64-encoded-hash"
		//	}
		field.JSON("disk_encryption_key_json", json.RawMessage{}).
			Optional(),

		// InitializeParamsJSON contains boot disk creation parameters.
		//
		//	{
		//	  "sourceImage": "projects/.../images/...",
		//	  "diskType": "pd-balanced" | "pd-ssd" | "pd-standard",
		//	  "diskSizeGb": "100"
		//	}
		field.JSON("initialize_params_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPComputeInstanceDisk) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("disks").
			Unique().
			Required(),
		edge.To("licenses", BronzeGCPComputeInstanceDiskLicense.Type),
	}
}

func (BronzeGCPComputeInstanceDisk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_disks"},
	}
}

// BronzeGCPComputeInstanceDiskLicense represents a software license on an attached disk.
// Data from instance.disks[].licenses[].
type BronzeGCPComputeInstanceDiskLicense struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceDiskLicense) Fields() []ent.Field {
	return []ent.Field{
		field.String("license").
			NotEmpty(),
	}
}

func (BronzeGCPComputeInstanceDiskLicense) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("disk", BronzeGCPComputeInstanceDisk.Type).
			Ref("licenses").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceDiskLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_disk_licenses"},
	}
}

// BronzeGCPComputeInstanceLabel represents a label on a GCP Compute instance.
// Data from instance.labels map.
type BronzeGCPComputeInstanceLabel struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_labels"},
	}
}

// BronzeGCPComputeInstanceMetadata represents instance metadata key-value pairs.
// Data from instance.metadata.items[].
type BronzeGCPComputeInstanceMetadata struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("metadata").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceMetadata) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_metadata"},
	}
}

// BronzeGCPComputeInstanceNIC represents a network interface on a GCP Compute instance.
// Data from instance.networkInterfaces[].
type BronzeGCPComputeInstanceNIC struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceNIC) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields (preserving original API field structure)
		field.String("name").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("network_ip").
			Optional(),
		field.String("stack_type").
			Optional(),
		field.String("nic_type").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceNIC) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("nics").
			Unique().
			Required(),
		edge.To("access_configs", BronzeGCPComputeInstanceNICAccessConfig.Type),
		edge.To("alias_ip_ranges", BronzeGCPComputeInstanceNICAliasRange.Type),
	}
}

func (BronzeGCPComputeInstanceNIC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nics"},
	}
}

// BronzeGCPComputeInstanceNICAccessConfig represents external IP configuration for a NIC.
// Data from instance.networkInterfaces[].accessConfigs[].
type BronzeGCPComputeInstanceNICAccessConfig struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceNICAccessConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("type").
			Optional(),
		field.String("name").
			Optional(),
		field.String("nat_ip").
			Optional(),
		field.String("network_tier").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceNICAccessConfig) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("nic", BronzeGCPComputeInstanceNIC.Type).
			Ref("access_configs").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceNICAccessConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nic_access_configs"},
	}
}

// BronzeGCPComputeInstanceNICAliasRange represents a secondary IP range on a NIC.
// Data from instance.networkInterfaces[].aliasIpRanges[].
type BronzeGCPComputeInstanceNICAliasRange struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceNICAliasRange) Fields() []ent.Field {
	return []ent.Field{
		field.String("ip_cidr_range").
			Optional(),
		field.String("subnetwork_range_name").
			Optional(),
	}
}

func (BronzeGCPComputeInstanceNICAliasRange) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("nic", BronzeGCPComputeInstanceNIC.Type).
			Ref("alias_ip_ranges").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceNICAliasRange) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nic_alias_ranges"},
	}
}

// BronzeGCPComputeInstanceServiceAccount represents a service account attached to an instance.
// Data from instance.serviceAccounts[].
type BronzeGCPComputeInstanceServiceAccount struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceServiceAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").
			NotEmpty(),

		// ScopesJSON contains OAuth scopes granted to the service account.
		//
		//	["https://www.googleapis.com/auth/cloud-platform", ...]
		field.JSON("scopes_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeGCPComputeInstanceServiceAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("service_accounts").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceServiceAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_service_accounts"},
	}
}

// BronzeGCPComputeInstanceTag represents a network tag on a GCP Compute instance.
// Data from instance.tags.items[].
type BronzeGCPComputeInstanceTag struct {
	ent.Schema
}

func (BronzeGCPComputeInstanceTag) Fields() []ent.Field {
	return []ent.Field{
		field.String("tag").
			NotEmpty(),
	}
}

func (BronzeGCPComputeInstanceTag) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("instance", BronzeGCPComputeInstance.Type).
			Ref("tags").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeInstanceTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_tags"},
	}
}
