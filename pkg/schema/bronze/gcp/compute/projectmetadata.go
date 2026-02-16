package compute

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

// BronzeGCPComputeProjectMetadata represents GCP Compute Engine project metadata in the bronze layer.
// Fields preserve raw API response data from compute.projects.get.
type BronzeGCPComputeProjectMetadata struct {
	ent.Schema
}

func (BronzeGCPComputeProjectMetadata) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeProjectMetadata) Fields() []ent.Field {
	return []ent.Field{
		// GCP API fields
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Project ID"),
		field.String("name").
			NotEmpty(),
		field.String("default_service_account").
			Optional(),
		field.String("default_network_tier").
			Optional(),
		field.String("xpn_project_status").
			Optional(),
		field.String("creation_timestamp").
			Optional(),

		// JSONB fields
		field.JSON("usage_export_location_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeProjectMetadata) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("items", BronzeGCPComputeProjectMetadataItem.Type),
	}
}

func (BronzeGCPComputeProjectMetadata) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeProjectMetadata) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_project_metadata"},
	}
}

// BronzeGCPComputeProjectMetadataItem represents a key-value metadata item in a GCP project.
// Data from project.commonInstanceMetadata.items[].
type BronzeGCPComputeProjectMetadataItem struct {
	ent.Schema
}

func (BronzeGCPComputeProjectMetadataItem) Fields() []ent.Field {
	return []ent.Field{
		field.String("key").
			NotEmpty(),
		field.Text("value").
			Optional(),
	}
}

func (BronzeGCPComputeProjectMetadataItem) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("metadata", BronzeGCPComputeProjectMetadata.Type).
			Ref("items").
			Unique().
			Required(),
	}
}

func (BronzeGCPComputeProjectMetadataItem) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_project_metadata_items"},
	}
}
