package volume

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

// BronzeGreenNodeVolumeBlockVolume represents a GreenNode block volume in the bronze layer.
type BronzeGreenNodeVolumeBlockVolume struct {
	ent.Schema
}

func (BronzeGreenNodeVolumeBlockVolume) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeVolumeBlockVolume) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Volume UUID"),
		field.String("name").
			NotEmpty(),
		field.String("volume_type_id").
			Optional(),
		field.String("cluster_id").
			Optional(),
		field.String("vm_id").
			Optional(),
		field.String("size").
			Optional().
			Comment("uint64 stored as string"),
		field.String("iops_id").
			Optional().
			Comment("uint64 stored as string"),
		field.String("status").
			Optional(),
		field.String("created_at_api").
			Optional().
			Comment("Volume creation timestamp from API"),
		field.String("updated_at_api").
			Optional().
			Comment("Volume update timestamp from API"),
		field.Bool("persistent_volume").
			Default(false),
		field.JSON("attached_machine_json", json.RawMessage{}).
			Optional().
			Comment("Attached machine IDs as JSON array"),
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
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeVolumeBlockVolume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("snapshots", BronzeGreenNodeVolumeSnapshot.Type),
	}
}

func (BronzeGreenNodeVolumeBlockVolume) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeVolumeBlockVolume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_block_volumes"},
	}
}

// BronzeGreenNodeVolumeSnapshot represents a snapshot attached to a block volume.
type BronzeGreenNodeVolumeSnapshot struct {
	ent.Schema
}

func (BronzeGreenNodeVolumeSnapshot) Fields() []ent.Field {
	return []ent.Field{
		field.String("snapshot_id").
			NotEmpty().
			Comment("Snapshot UUID from SDK"),
		field.String("name").
			Optional(),
		field.Int64("size").
			Optional(),
		field.Int64("volume_size").
			Optional(),
		field.String("status").
			Optional(),
		field.String("created_at_api").
			Optional().
			Comment("Snapshot creation timestamp from API"),
	}
}

func (BronzeGreenNodeVolumeSnapshot) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("block_volume", BronzeGreenNodeVolumeBlockVolume.Type).
			Ref("snapshots").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeVolumeSnapshot) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_volume_snapshots"},
	}
}
