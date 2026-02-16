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

// BronzeHistoryGCPAppEngineApplication stores historical snapshots of GCP App Engine applications.
type BronzeHistoryGCPAppEngineApplication struct {
	ent.Schema
}

func (BronzeHistoryGCPAppEngineApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPAppEngineApplication) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze application by resource_id"),

		// All application fields
		field.String("name").
			NotEmpty(),
		field.String("auth_domain").
			Optional(),
		field.String("location_id").
			Optional(),
		field.String("code_bucket").
			Optional(),
		field.String("default_cookie_expiration").
			Optional(),
		field.Int32("serving_status").
			Optional(),
		field.String("default_hostname").
			Optional(),
		field.String("default_bucket").
			Optional(),
		field.String("gcr_domain").
			Optional(),
		field.Int32("database_type").
			Optional(),

		// JSONB fields
		field.JSON("feature_settings_json", json.RawMessage{}).
			Optional(),
		field.JSON("iap_json", json.RawMessage{}).
			Optional(),
		field.JSON("dispatch_rules_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPAppEngineApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPAppEngineApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_appengine_applications_history"},
	}
}
