package serviceusage

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPServiceUsageEnabledService stores historical snapshots of GCP enabled services.
type BronzeHistoryGCPServiceUsageEnabledService struct {
	ent.Schema
}

func (BronzeHistoryGCPServiceUsageEnabledService) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPServiceUsageEnabledService) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze enabled service by resource_id"),

		// All enabled service fields
		field.String("name").
			NotEmpty().
			Comment("Service resource name"),
		field.String("parent").
			NotEmpty().
			Comment("Parent resource (projects/{project})"),

		// JSONB fields
		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("ServiceConfig as JSON"),

		field.Int("state").
			Default(0).
			Comment("Service state: 0=STATE_UNSPECIFIED, 1=DISABLED, 2=ENABLED"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPServiceUsageEnabledService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPServiceUsageEnabledService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_serviceusage_enabled_services_history"},
	}
}
