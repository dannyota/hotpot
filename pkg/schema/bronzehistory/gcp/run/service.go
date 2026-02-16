package run

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPRunService stores historical snapshots of GCP Cloud Run services.
type BronzeHistoryGCPRunService struct {
	ent.Schema
}

func (BronzeHistoryGCPRunService) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPRunService) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze service by resource_id"),

		// All service fields
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("uid").
			Optional(),
		field.Int64("generation").
			Optional(),

		// JSONB fields
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("annotations_json", json.RawMessage{}).
			Optional(),

		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
		field.String("creator").
			Optional(),
		field.String("last_modifier").
			Optional(),
		field.Int("ingress").
			Optional().
			Comment("IngressTraffic enum value"),
		field.Int("launch_stage").
			Optional().
			Comment("LaunchStage enum value"),
		field.JSON("template_json", json.RawMessage{}).
			Optional(),
		field.JSON("traffic_json", json.RawMessage{}).
			Optional(),
		field.String("uri").
			Optional(),
		field.Int64("observed_generation").
			Optional(),
		field.JSON("terminal_condition_json", json.RawMessage{}).
			Optional(),
		field.JSON("conditions_json", json.RawMessage{}).
			Optional(),
		field.String("latest_ready_revision").
			Optional(),
		field.String("latest_created_revision").
			Optional(),
		field.JSON("traffic_statuses_json", json.RawMessage{}).
			Optional(),
		field.Bool("reconciling").
			Default(false),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPRunService) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPRunService) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_run_services_history"},
	}
}
