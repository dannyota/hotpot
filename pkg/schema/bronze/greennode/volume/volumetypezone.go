package volume

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeVolumeVolumeTypeZone represents a GreenNode volume type zone in the bronze layer.
type BronzeGreenNodeVolumeVolumeTypeZone struct {
	ent.Schema
}

func (BronzeGreenNodeVolumeVolumeTypeZone) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeVolumeVolumeTypeZone) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Volume type zone UUID"),
		field.String("name").
			NotEmpty(),
		field.JSON("pool_name_json", json.RawMessage{}).
			Optional().
			Comment("Pool names as JSON array"),

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeVolumeVolumeTypeZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeVolumeVolumeTypeZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_volume_type_zones"},
	}
}
