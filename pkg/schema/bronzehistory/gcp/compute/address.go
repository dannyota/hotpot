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

// BronzeHistoryGCPComputeAddress stores historical snapshots of GCP Compute regional addresses.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeAddress struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze address by resource_id"),

		// All address fields (same as bronze.BronzeGCPComputeAddress)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("address").
			Optional(),
		field.String("address_type").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.String("ipv6_endpoint_type").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("region").
			Optional(),
		field.String("status").
			Optional(),
		field.String("purpose").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Int32("prefix_length").
			Default(0),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// JSON arrays
		field.JSON("users_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_addresses_history"},
	}
}

// BronzeHistoryGCPComputeAddressLabel stores historical snapshots of address labels.
// Links via address_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeAddressLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeAddressLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("address_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeAddress"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPComputeAddressLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("address_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeAddressLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_address_labels_history"},
	}
}

// BronzeHistoryGCPComputeGlobalAddress stores historical snapshots of GCP Compute global addresses.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeGlobalAddress struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeGlobalAddress) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeGlobalAddress) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze global address by resource_id"),

		// All address fields (same as bronze.BronzeGCPComputeGlobalAddress)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("address").
			Optional(),
		field.String("address_type").
			Optional(),
		field.String("ip_version").
			Optional(),
		field.String("ipv6_endpoint_type").
			Optional(),
		field.String("ip_collection").
			Optional(),
		field.String("region").
			Optional(),
		field.String("status").
			Optional(),
		field.String("purpose").
			Optional(),
		field.String("network").
			Optional(),
		field.String("subnetwork").
			Optional(),
		field.String("network_tier").
			Optional(),
		field.Int32("prefix_length").
			Default(0),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("label_fingerprint").
			Optional(),

		// JSON arrays
		field.JSON("users_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeGlobalAddress) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeGlobalAddress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_global_addresses_history"},
	}
}

// BronzeHistoryGCPComputeGlobalAddressLabel stores historical snapshots of global address labels.
// Links via global_address_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeGlobalAddressLabel struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeGlobalAddressLabel) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("global_address_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeGlobalAddress"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Label fields
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeHistoryGCPComputeGlobalAddressLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("global_address_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeGlobalAddressLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		// _history suffix: workaround for ent Issue #2330
		entsql.Annotation{Table: "gcp_compute_global_address_labels_history"},
	}
}
