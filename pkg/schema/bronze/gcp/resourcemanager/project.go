package resourcemanager

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPProject represents a GCP project in the bronze layer.
// Fields preserve raw API response data from cloudresourcemanager.projects.search.
type BronzeGCPProject struct {
	ent.Schema
}

func (BronzeGCPProject) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPProject) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("project_id").
			Unique().
			Immutable().
			Comment("User-assigned project ID (e.g., 'my-project-123')"),
		field.String("project_number").
			NotEmpty().
			Unique().
			Comment("Google-assigned project number"),
		field.String("display_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("parent").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
		field.String("etag").
			Optional(),
	}
}

func (BronzeGCPProject) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPProjectLabel.Type),
	}
}

func (BronzeGCPProject) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("state"),
		index.Fields("parent"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPProject) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_projects"},
	}
}

// BronzeGCPProjectLabel represents a label attached to a GCP project.
type BronzeGCPProjectLabel struct {
	ent.Schema
}

func (BronzeGCPProjectLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPProjectLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", BronzeGCPProject.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPProjectLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_project_labels"},
	}
}
