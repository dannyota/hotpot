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

// BronzeHistoryDODatabaseConfig stores historical snapshots of DigitalOcean Database Configs.
type BronzeHistoryDODatabaseConfig struct {
	ent.Schema
}

func (BronzeHistoryDODatabaseConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabaseConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabaseConfig by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("engine_slug").
			Optional(),
		field.JSON("config_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryDODatabaseConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("engine_slug"),
	}
}

func (BronzeHistoryDODatabaseConfig) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_configs_history"},
	}
}
