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

// BronzeGCPFolder represents a GCP folder in the bronze layer.
// Fields preserve raw API response data from cloudresourcemanager.folders.list.
type BronzeGCPFolder struct {
	ent.Schema
}

func (BronzeGCPFolder) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPFolder) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Folder resource name (e.g., folders/123)"),
		field.String("name").
			NotEmpty(),
		field.String("display_name").
			Optional(),
		field.String("state").
			Optional(),
		field.String("parent").
			Optional(),
		field.String("etag").
			Optional(),
		field.String("create_time").
			Optional(),
		field.String("update_time").
			Optional(),
		field.String("delete_time").
			Optional(),
	}
}

func (BronzeGCPFolder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("labels", BronzeGCPFolderLabel.Type),
	}
}

func (BronzeGCPFolder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("state"),
		index.Fields("parent"),
	}
}

func (BronzeGCPFolder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_folders"},
	}
}

// BronzeGCPFolderLabel represents a label attached to a GCP folder.
type BronzeGCPFolderLabel struct {
	ent.Schema
}

func (BronzeGCPFolderLabel) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.String("value"),
	}
}

func (BronzeGCPFolderLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("folder", BronzeGCPFolder.Type).
			Ref("labels").
			Unique().
			Required(),
	}
}

func (BronzeGCPFolderLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_folder_labels"},
	}
}
