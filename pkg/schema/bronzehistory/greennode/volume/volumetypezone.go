package volume

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeVolumeVolumeTypeZone stores historical snapshots of GreenNode volume type zones.
type BronzeHistoryGreenNodeVolumeVolumeTypeZone struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeVolumeVolumeTypeZone) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeVolumeVolumeTypeZone) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.JSON("pool_name_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeVolumeVolumeTypeZone) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeVolumeVolumeTypeZone) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_volume_type_zones_history"},
	}
}
