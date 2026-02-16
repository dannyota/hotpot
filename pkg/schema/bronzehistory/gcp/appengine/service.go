package appengine

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPAppEngineService stores historical snapshots of GCP App Engine services.
type BronzeHistoryGCPAppEngineService struct {
	ent.Schema
}

func (BronzeHistoryGCPAppEngineService) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAppEngineService) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze service by resource_id"),
		field.Uint("application_history_id").
			Comment("Links to parent BronzeHistoryGCPAppEngineApplication"),

		// All service fields
		field.String("name").
			NotEmpty(),

		// JSONB fields
		field.JSON("split_json", json.RawMessage{}).
			Optional(),
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("network_settings_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAppEngineService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("application_history_id"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPAppEngineService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_appengine_services_history"},
	}
}
