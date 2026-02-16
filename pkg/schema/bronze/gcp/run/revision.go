package run

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPRunRevision represents a GCP Cloud Run revision in the bronze layer.
// Fields preserve raw API response data from run.revisions.list.
type BronzeGCPRunRevision struct {
	ent.Schema
}

func (BronzeGCPRunRevision) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPRunRevision) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Revision resource name (projects/{project}/locations/{location}/services/{service}/revisions/{revision})"),
		field.String("name").
			NotEmpty(),
		field.String("uid").
			Optional(),
		field.Int64("generation").
			Optional(),

		// LabelsJSON contains labels inherited from the service and revision template.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// AnnotationsJSON contains annotations inherited from the service and revision template.
		//
		//	{"key1": "value1", "key2": "value2"}
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

		// ScalingJSON contains revision scaling configuration.
		//
		//	{"minInstanceCount": 0, "maxInstanceCount": 100}
		field.JSON("scaling_json", json.RawMessage{}).
			Optional(),

		// ContainersJSON contains the container specifications.
		//
		//	[{"image": "...", "ports": [...], "env": [...], "resources": {...}}]
		field.JSON("containers_json", json.RawMessage{}).
			Optional(),

		// VolumesJSON contains volume definitions.
		//
		//	[{"name": "...", "secret": {...}}]
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

		// ConditionsJSON contains revision readiness conditions.
		//
		//	[{"type": "...", "state": "...", "message": "..."}]
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

func (BronzeGCPRunRevision) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("service", BronzeGCPRunService.Type).
			Ref("revisions").
			Unique().
			Required(),
	}
}

func (BronzeGCPRunRevision) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("location"),
		index.Fields("service_name"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPRunRevision) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_run_revisions"},
	}
}
