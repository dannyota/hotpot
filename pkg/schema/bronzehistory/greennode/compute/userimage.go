package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeComputeUserImage stores historical snapshots of GreenNode user images.
type BronzeHistoryGreenNodeComputeUserImage struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeUserImage) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeComputeUserImage) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.Int("min_disk").
			Optional(),
		field.Float("image_size").
			Optional(),
		field.String("meta_data").
			Optional(),
		field.String("created_at").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeUserImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeComputeUserImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_user_images_history"},
	}
}
