package volume

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeVolumeVolumeType represents a GreenNode volume type in the bronze layer.
type BronzeGreenNodeVolumeVolumeType struct {
	ent.Schema
}

func (BronzeGreenNodeVolumeVolumeType) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeVolumeVolumeType) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Volume type UUID"),
		field.String("name").
			NotEmpty(),
		field.Int("iops").
			Optional(),
		field.Int("max_size").
			Optional(),
		field.Int("min_size").
			Optional(),
		field.Int("through_put").
			Optional(),
		field.String("zone_id").
			Optional(),

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeVolumeVolumeType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeVolumeVolumeType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_volume_types"},
	}
}
