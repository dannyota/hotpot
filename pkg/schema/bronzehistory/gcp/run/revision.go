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

// BronzeHistoryGCPRunRevision stores historical snapshots of GCP Cloud Run revisions.
type BronzeHistoryGCPRunRevision struct {
	ent.Schema
}

func (BronzeHistoryGCPRunRevision) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPRunRevision) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze revision by resource_id"),

		// All revision fields
		field.String("name").
			NotEmpty(),
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
		field.Int("launch_stage").
			Optional().
			Comment("LaunchStage enum value"),
		field.String("service_name").
			Optional().
			Comment("Parent service resource name"),
		field.JSON("scaling_json", json.RawMessage{}).
			Optional(),
		field.JSON("containers_json", json.RawMessage{}).
			Optional(),
		field.JSON("volumes_json", json.RawMessage{}).
			Optional(),
		field.Int("execution_environment").
			Optional().
			Comment("ExecutionEnvironment enum value"),
		field.String("encryption_key").
			Optional(),
		field.Int("max_instance_request_concurrency").
			Optional(),
		field.String("timeout").
			Optional().
			Comment("Duration string (e.g. '300s')"),
		field.String("service_account").
			Optional(),
		field.Bool("reconciling").
			Default(false),
		field.JSON("conditions_json", json.RawMessage{}).
			Optional(),
		field.Int64("observed_generation").
			Optional(),
		field.String("log_uri").
			Optional(),
		field.String("etag").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
		field.String("location").
			NotEmpty(),
	}
}

func (BronzeHistoryGCPRunRevision) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGCPRunRevision) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_run_revisions_history"},
	}
}
