package volume

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeVolumeBlockVolume stores historical snapshots of GreenNode block volumes.
type BronzeHistoryGreenNodeVolumeBlockVolume struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeVolumeBlockVolume) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeVolumeBlockVolume) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("volume_type_id").
			Optional(),
		field.String("cluster_id").
			Optional(),
		field.String("vm_id").
			Optional(),
		field.String("size").
			Optional(),
		field.String("iops_id").
			Optional(),
		field.String("status").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.Bool("persistent_volume").
			Default(false),
		field.JSON("attached_machine_json", json.RawMessage{}).
			Optional(),
		field.String("under_id").
			Optional(),
		field.String("migrate_state").
			Optional(),
		field.Bool("multi_attach").
			Default(false),
		field.String("zone_id").
			Optional(),

		// Collection metadata
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeVolumeBlockVolume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeVolumeBlockVolume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_block_volumes_history"},
	}
}

// BronzeHistoryGreenNodeVolumeSnapshot stores historical snapshots of volume snapshots.
type BronzeHistoryGreenNodeVolumeSnapshot struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeVolumeSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("block_volume_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeVolumeBlockVolume"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("snapshot_id").
			NotEmpty(),
		field.String("name").
			Optional(),
		field.Int64("size").
			Optional(),
		field.Int64("volume_size").
			Optional(),
		field.String("status").
			Optional(),
		field.String("created_at_api").
			Optional(),
	}
}

func (BronzeHistoryGreenNodeVolumeSnapshot) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("block_volume_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeVolumeSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_snapshots_history"},
	}
}
