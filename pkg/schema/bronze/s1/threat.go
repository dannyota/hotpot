package s1

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1Threat represents a SentinelOne threat in the bronze layer.
type BronzeS1Threat struct {
	ent.Schema
}

func (BronzeS1Threat) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1Threat) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne threat ID"),
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

		// ThreatInfoJSON stores the full threat info object from the API.
		//
		//	{"sha256": "...", "md5": "...", "filePath": "...", ...}
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

func (BronzeS1Threat) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id"),
		index.Fields("status"),
		index.Fields("classification"),
		index.Fields("collected_at"),
		index.Fields("site_id"),
		index.Fields("account_id"),
	}
}

func (BronzeS1Threat) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_threats"},
	}
}
