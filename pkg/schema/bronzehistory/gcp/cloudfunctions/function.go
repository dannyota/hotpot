package cloudfunctions

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPCloudFunctionsFunction stores historical snapshots of GCP Cloud Functions.
type BronzeHistoryGCPCloudFunctionsFunction struct {
	ent.Schema
}

func (BronzeHistoryGCPCloudFunctionsFunction) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPCloudFunctionsFunction) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze function by resource_id"),

		// All function fields
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Int("environment").
			Default(0).
			Comment("Function environment (0=UNSPECIFIED, 1=GEN_1, 2=GEN_2)"),
		field.Int("state").
			Default(0).
			Comment("Function state (0=UNSPECIFIED, 1=ACTIVE, 2=FAILED, 3=DEPLOYING, 4=DELETING, 5=UNKNOWN)"),

		// JSONB fields
		field.JSON("build_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("service_config_json", json.RawMessage{}).
			Optional(),
		field.JSON("event_trigger_json", json.RawMessage{}).
			Optional(),
		field.JSON("state_messages_json", json.RawMessage{}).
			Optional(),

		field.String("update_time").
			Optional(),
		field.String("create_time").
			Optional(),

		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		field.String("kms_key_name").
			Optional(),
		field.String("url").
			Optional(),
		field.Bool("satisfies_pzs").
			Default(false),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			Optional(),
	}
}

func (BronzeHistoryGCPCloudFunctionsFunction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPCloudFunctionsFunction) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudfunctions_functions_history"},
	}
}
