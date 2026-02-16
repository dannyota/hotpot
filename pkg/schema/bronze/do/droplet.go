package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODroplet represents a DigitalOcean Droplet in the bronze layer.
type BronzeDODroplet struct {
	ent.Schema
}

func (BronzeDODroplet) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODroplet) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Droplet ID (int converted to string)"),
		field.String("name").
			NotEmpty(),
		field.Int("memory").
			Default(0),
		field.Int("vcpus").
			Default(0),
		field.Int("disk").
			Default(0),
		field.String("region").
			Optional().
			Comment("Region slug"),
		field.String("size_slug").
			Optional(),
		field.String("status").
			Optional(),
		field.Bool("locked").
			Default(false),
		field.String("vpc_uuid").
			Optional(),
		field.String("api_created_at").
			Optional().
			Comment("API-reported creation timestamp"),
		field.JSON("image_json", json.RawMessage{}).
			Optional().
			Comment("Raw image object JSON"),
		field.JSON("size_json", json.RawMessage{}).
			Optional().
			Comment("Raw size object JSON"),
		field.JSON("networks_json", json.RawMessage{}).
			Optional().
			Comment("Raw networks object JSON"),
		field.JSON("kernel_json", json.RawMessage{}).
			Optional().
			Comment("Raw kernel object JSON"),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("features_json", json.RawMessage{}).
			Optional(),
		field.JSON("volume_ids_json", json.RawMessage{}).
			Optional(),
		field.JSON("backup_ids_json", json.RawMessage{}).
			Optional(),
		field.JSON("snapshot_ids_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeDODroplet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region"),
		index.Fields("status"),
		index.Fields("vpc_uuid"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODroplet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_droplets"},
	}
}
