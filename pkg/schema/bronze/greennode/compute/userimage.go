package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeComputeUserImage represents a GreenNode user image in the bronze layer.
type BronzeGreenNodeComputeUserImage struct {
	ent.Schema
}

func (BronzeGreenNodeComputeUserImage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeComputeUserImage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("User image UUID"),
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
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeUserImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeComputeUserImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_user_images"},
	}
}
