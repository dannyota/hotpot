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

// BronzeDODatabaseConfig represents engine-specific configuration for a DigitalOcean Database in the bronze layer.
type BronzeDODatabaseConfig struct {
	ent.Schema
}

func (BronzeDODatabaseConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabaseConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Cluster ID (1:1 with cluster)"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("engine_slug").
			Optional(),
		field.JSON("config_json", json.RawMessage{}).
			Optional().
			Comment("Engine-specific configuration"),
	}
}

func (BronzeDODatabaseConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("engine_slug"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabaseConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_configs"},
	}
}
