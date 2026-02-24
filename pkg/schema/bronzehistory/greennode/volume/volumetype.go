package volume

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeVolumeVolumeType stores historical snapshots of GreenNode volume types.
type BronzeHistoryGreenNodeVolumeVolumeType struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeVolumeVolumeType) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeVolumeVolumeType) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeVolumeVolumeType) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeVolumeVolumeType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_volume_types_history"},
	}
}
