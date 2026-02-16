package do

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODroplet stores historical snapshots of DigitalOcean Droplets.
type BronzeHistoryDODroplet struct {
	ent.Schema
}

func (BronzeHistoryDODroplet) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODroplet) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Droplet by resource_id"),
		field.String("name").
			NotEmpty(),
		field.Int("memory").
			Default(0),
		field.Int("vcpus").
			Default(0),
		field.Int("disk").
			Default(0),
		field.String("region").
			Optional(),
		field.String("size_slug").
			Optional(),
		field.String("status").
			Optional(),
		field.Bool("locked").
			Default(false),
		field.String("vpc_uuid").
			Optional(),
		field.String("api_created_at").
			Optional(),
		field.JSON("image_json", json.RawMessage{}).
			Optional(),
		field.JSON("size_json", json.RawMessage{}).
			Optional(),
		field.JSON("networks_json", json.RawMessage{}).
			Optional(),
		field.JSON("kernel_json", json.RawMessage{}).
			Optional(),
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

func (BronzeHistoryDODroplet) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("region"),
		index.Fields("status"),
		index.Fields("vpc_uuid"),
	}
}

func (BronzeHistoryDODroplet) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_droplets_history"},
	}
}
