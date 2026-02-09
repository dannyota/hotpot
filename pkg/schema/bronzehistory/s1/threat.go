package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1Threat stores historical snapshots of SentinelOne threats.
type BronzeHistoryS1Threat struct {
	ent.Schema
}

func (BronzeHistoryS1Threat) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1Threat) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze threat by resource_id"),

		field.String("agent_id").
			Optional(),
		field.String("classification").
			Optional(),
		field.String("threat_name").
			Optional(),
		field.String("file_path").
			Optional(),
		field.String("status").
			Optional(),
		field.String("analyst_verdict").
			Optional(),
		field.String("confidence_level").
			Optional(),
		field.String("initiated_by").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.JSON("threat_info_json", json.RawMessage{}).
			Optional(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
		field.String("file_content_hash").
			Optional(),
		field.String("file_sha256").
			Optional(),
		field.String("cloud_verdict").
			Optional(),
		field.String("classification_source").
			Optional(),
		field.String("site_id").
			Optional(),
		field.String("site_name").
			Optional(),
		field.String("account_id").
			Optional(),
		field.String("account_name").
			Optional(),
		field.String("agent_computer_name").
			Optional(),
		field.String("agent_os_type").
			Optional(),
		field.String("agent_machine_type").
			Optional(),
		field.Bool("agent_is_active").
			Default(false),
		field.Bool("agent_is_decommissioned").
			Default(false),
		field.String("agent_version").
			Optional(),
	}
}

func (BronzeHistoryS1Threat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("agent_id"),
	}
}

func (BronzeHistoryS1Threat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_threats_history"},
	}
}
