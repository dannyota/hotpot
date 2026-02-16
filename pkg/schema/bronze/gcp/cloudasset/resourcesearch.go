package cloudasset

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPCloudAssetResourceSearch represents a resource search result in the bronze layer.
// Fields preserve raw API response data from asset.SearchAllResources.
type BronzeGCPCloudAssetResourceSearch struct {
	ent.Schema
}

func (BronzeGCPCloudAssetResourceSearch) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPCloudAssetResourceSearch) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Full resource name (e.g., //compute.googleapis.com/projects/123/zones/zone1/instances/instance1)"),
		field.String("asset_type").
			NotEmpty().
			Comment("Asset type (e.g., compute.googleapis.com/Disk)"),
		field.String("project").
			Optional().
			Comment("Project that the resource belongs to (projects/{PROJECT_NUMBER})"),
		field.String("display_name").
			Optional().
			Comment("Display name of the resource"),
		field.String("description").
			Optional().
			Comment("Text description of the resource"),
		field.String("location").
			Optional().
			Comment("Resource location (global, regional, or zonal)"),
		field.String("organization_id").
			NotEmpty().
			Comment("Organization resource name (scope of the search)"),

		// JSON fields for nested data
		field.JSON("labels_json", json.RawMessage{}).
			Optional().
			Comment("User labels associated with the resource as JSON"),
		field.JSON("network_tags_json", json.RawMessage{}).
			Optional().
			Comment("Network tags associated with the resource as JSON array"),
		field.JSON("additional_attributes_json", json.RawMessage{}).
			Optional().
			Comment("Additional searchable attributes as JSON"),
	}
}

func (BronzeGCPCloudAssetResourceSearch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("collected_at"),
		index.Fields("organization_id"),
		index.Fields("asset_type"),
		index.Fields("project"),
		index.Fields("location"),
	}
}

func (BronzeGCPCloudAssetResourceSearch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_resource_searches"},
	}
}
