package dns

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPDNSPolicy stores historical snapshots of GCP Cloud DNS policies.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPDNSPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPDNSPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPDNSPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DNS policy by resource_id"),

		// All policy fields (same as bronze.BronzeGCPDNSPolicy)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Bool("enable_inbound_forwarding").
			Default(false),
		field.Bool("enable_logging").
			Default(false),

		// JSONB fields
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),
		field.JSON("alternative_name_server_config_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPDNSPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPDNSPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_dns_policies_history"},
	}
}
