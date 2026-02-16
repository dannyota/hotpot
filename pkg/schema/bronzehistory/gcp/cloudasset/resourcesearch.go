package cloudasset

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGCPCloudAssetResourceSearch stores historical snapshots of resource search results.
// Uses resource_id for lookup, with valid_from/valid_to for time range.
type BronzeHistoryGCPCloudAssetResourceSearch struct {
	ent.Schema
}

func (BronzeHistoryGCPCloudAssetResourceSearch) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPCloudAssetResourceSearch) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze resource search by resource_id"),

		// All resource search fields (same as bronze.BronzeGCPCloudAssetResourceSearch)
		field.String("asset_type").
			NotEmpty(),
		field.String("project").
			Optional(),
		field.String("display_name").
			Optional(),
		field.String("description").
			Optional(),
		field.String("location").
			Optional(),
		field.String("organization_id").
			NotEmpty(),

		// JSON fields for nested data
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),
		field.JSON("network_tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("additional_attributes_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryGCPCloudAssetResourceSearch) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPCloudAssetResourceSearch) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_cloudasset_resource_searches_history"},
	}
}
