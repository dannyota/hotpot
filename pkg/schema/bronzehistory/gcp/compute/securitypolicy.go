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

// BronzeHistoryGCPComputeSecurityPolicy stores historical snapshots of GCP Cloud Armor security policies.
// Uses resource_id for lookup (has API ID), with valid_from/valid_to for time range.
type BronzeHistoryGCPComputeSecurityPolicy struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeSecurityPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeSecurityPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze security policy by resource_id"),

		// All security policy fields (same as bronze.BronzeGCPComputeSecurityPolicy)
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// Security policy configuration
		field.String("type").
			Optional(),
		field.String("fingerprint").
			Optional(),

		// JSONB fields
		field.JSON("rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("associations_json", json.RawMessage{}).
			Optional(),
		field.JSON("adaptive_protection_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("advanced_options_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("ddos_protection_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("recaptcha_options_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeSecurityPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPComputeSecurityPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_security_policies_history"},
	}
}
