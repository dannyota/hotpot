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

// BronzeHistoryGCPComputeSubnetwork stores historical snapshots of GCP VPC subnetworks.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeSubnetwork struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSubnetwork) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeSubnetwork) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze subnetwork by resource_id"),

		// All subnetwork fields (same as bronze.BronzeGCPComputeSubnetwork)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Network relationship
		field.String("network").
			NotEmpty(),
		field.String("region").
			NotEmpty(),

		// IP configuration
		field.String("ip_cidr_range").
			NotEmpty(),
		field.String("gateway_address").
			Optional(),

		// Purpose and role
		field.String("purpose").
			Optional(),
		field.String("role").
			Optional(),

		// Private Google Access
		field.Bool("private_ip_google_access").
			Default(false),
		field.String("private_ipv6_google_access").
			Optional(),

		// Stack type and IPv6
		field.String("stack_type").
			Optional(),
		field.String("ipv6_access_type").
			Optional(),
		field.String("internal_ipv6_prefix").
			Optional(),
		field.String("external_ipv6_prefix").
			Optional(),

		// LogConfigJSON
		field.JSON("log_config_json", json.RawMessage{}).
			Optional(),

		// Fingerprint
		field.String("fingerprint").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeSubnetwork) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeSubnetwork) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_subnetworks_history"},
	}
}

// BronzeHistoryGCPComputeSubnetworkSecondaryRange stores historical snapshots of secondary IP ranges.
// Links via subnetwork_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeSubnetworkSecondaryRange struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSubnetworkSecondaryRange) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("subnetwork_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeSubnetwork"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// All secondary range fields
		field.String("range_name").
			NotEmpty(),
		field.String("ip_cidr_range").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeSubnetworkSecondaryRange) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subnetwork_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeSubnetworkSecondaryRange) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_subnetwork_secondary_ranges_history"},
	}
}
