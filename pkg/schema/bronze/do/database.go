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

// BronzeDODatabase represents a DigitalOcean managed database cluster in the bronze layer.
type BronzeDODatabase struct {
	ent.Schema
}

func (BronzeDODatabase) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabase) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Database Cluster UUID"),
		field.String("name").
			Optional(),
		field.String("engine_slug").
			Optional().
			Comment("Database engine (pg, mysql, redis, mongodb, kafka, opensearch, valkey)"),
		field.String("version_slug").
			Optional(),
		field.Int("num_nodes").
			Default(0),
		field.String("size_slug").
			Optional(),
		field.String("region_slug").
			Optional(),
		field.String("status").
			Optional(),
		field.String("project_id").
			Optional(),
		field.Uint64("storage_size_mib").
			Default(0),
		field.String("private_network_uuid").
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.JSON("maintenance_window_json", json.RawMessage{}).
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDODatabase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("engine_slug"),
		index.Fields("region_slug"),
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabase) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_databases"},
	}
}
