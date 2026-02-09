package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPComputeInstance stores historical snapshots of GCP Compute instances.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeInstance struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeInstance) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze instance by resource_id"),

		// All instance fields (same as bronze.BronzeGCPComputeInstance)
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
		field.JSON("scheduling_json", json.RawMessage{}).
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instances_history"},
	}
}

// BronzeHistoryGCPComputeInstanceDisk stores historical snapshots of instance attached disks.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceDisk struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceDisk) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All disk fields (same as bronze.BronzeGCPComputeInstanceDisk)
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
		field.JSON("disk_encryption_key_json", json.RawMessage{}).
			Optional(),
		field.JSON("initialize_params_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceDisk) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceDisk) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_disks_history"},
	}
}

// BronzeHistoryGCPComputeInstanceDiskLicense stores historical snapshots of disk licenses.
// Links via disk_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceDiskLicense struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceDiskLicense) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("disk_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstanceDisk"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("license").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceDiskLicense) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("disk_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceDiskLicense) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_disk_licenses_history"},
	}
}

// BronzeHistoryGCPComputeInstanceLabel stores historical snapshots of instance labels.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("key").
			Optional(),
		field.String("value").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_labels_history"},
	}
}

// BronzeHistoryGCPComputeInstanceMetadata stores historical snapshots of instance metadata.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceMetadata struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceMetadata) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("key").
			Optional(),
		field.String("value").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceMetadata) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_metadata_history"},
	}
}

// BronzeHistoryGCPComputeInstanceNIC stores historical snapshots of instance network interfaces.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceNIC struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceNIC) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All NIC fields (same as bronze.BronzeGCPComputeInstanceNIC)
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

func (BronzeHistoryGCPComputeInstanceNIC) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceNIC) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nics_history"},
	}
}

// BronzeHistoryGCPComputeInstanceNICAccessConfig stores historical snapshots of NIC access configs.
// Links via nic_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceNICAccessConfig struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceNICAccessConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("nic_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstanceNIC"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All access config fields
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

func (BronzeHistoryGCPComputeInstanceNICAccessConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("nic_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceNICAccessConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nic_access_configs_history"},
	}
}

// BronzeHistoryGCPComputeInstanceNICAliasRange stores historical snapshots of NIC alias IP ranges.
// Links via nic_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceNICAliasRange struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceNICAliasRange) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("nic_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstanceNIC"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("ip_cidr_range").
			Optional(),
		field.String("subnetwork_range_name").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceNICAliasRange) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("nic_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceNICAliasRange) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_nic_alias_ranges_history"},
	}
}

// BronzeHistoryGCPComputeInstanceServiceAccount stores historical snapshots of instance service accounts.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceServiceAccount struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceServiceAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("email").
			Optional(),
		field.JSON("scopes_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceServiceAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceServiceAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_service_accounts_history"},
	}
}

// BronzeHistoryGCPComputeInstanceTag stores historical snapshots of instance network tags.
// Links via instance_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeInstanceTag struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeInstanceTag) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("instance_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeInstance"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		field.String("tag").
			Optional(),
	}
}

func (BronzeHistoryGCPComputeInstanceTag) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("instance_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeInstanceTag) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_instance_tags_history"},
	}
}
