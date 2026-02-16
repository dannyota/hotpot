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

// BronzeHistoryGCPComputeFirewall stores historical snapshots of GCP Compute firewall rules.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeFirewall struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeFirewall) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeFirewall) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze firewall by resource_id"),

		// All firewall fields (same as bronze.BronzeGCPComputeFirewall)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Firewall configuration
		field.String("network").
			Optional(),
		field.Int32("priority").
			Default(1000),
		field.String("direction").
			Optional(),
		field.Bool("disabled").
			Default(false),

		// JSONB fields
		field.JSON("source_ranges_json", json.RawMessage{}).
			Optional(),
		field.JSON("destination_ranges_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("target_tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("source_service_accounts_json", json.RawMessage{}).
			Optional(),
		field.JSON("target_service_accounts_json", json.RawMessage{}).
			Optional(),
		field.JSON("log_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeFirewall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeFirewall) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewalls_history"},
	}
}

// BronzeHistoryGCPComputeFirewallAllowed stores historical snapshots of firewall allowed rules.
// Links via firewall_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeFirewallAllowed struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeFirewallAllowed) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("firewall_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeFirewall"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Allowed rule fields
		field.String("ip_protocol").
			NotEmpty(),
		field.JSON("ports_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPComputeFirewallAllowed) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("firewall_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeFirewallAllowed) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewall_alloweds_history"},
	}
}

// BronzeHistoryGCPComputeFirewallDenied stores historical snapshots of firewall denied rules.
// Links via firewall_history_id, has own valid_from/valid_to for granular tracking.
type BronzeHistoryGCPComputeFirewallDenied struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeFirewallDenied) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.Uint("firewall_history_id").
			Comment("Links to parent BronzeHistoryGCPComputeFirewall"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),

		// Denied rule fields
		field.String("ip_protocol").
			NotEmpty(),
		field.JSON("ports_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPComputeFirewallDenied) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("firewall_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGCPComputeFirewallDenied) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_firewall_denieds_history"},
	}
}
