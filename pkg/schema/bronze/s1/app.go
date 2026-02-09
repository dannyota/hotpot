package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1App represents a SentinelOne installed application in the bronze layer.
type BronzeS1App struct {
	ent.Schema
}

func (BronzeS1App) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1App) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SentinelOne installed application ID"),
		field.String("name").
			NotEmpty(),
		field.String("publisher").
			Optional(),
		field.String("version").
			Optional(),
		field.Int64("size").
			Default(0),
		field.String("app_type").
			Optional(),
		field.String("os_type").
			Optional(),
		field.Time("installed_date").
			Optional().
			Nillable(),
		field.String("agent_id").
			Optional(),
		field.String("agent_computer_name").
			Optional(),
		field.String("agent_machine_type").
			Optional(),
		field.Bool("agent_is_active").
			Default(false),
		field.Bool("agent_is_decommissioned").
			Default(false),
		field.String("risk_level").
			Optional(),
		field.Bool("signed").
			Default(false),
		field.Time("api_created_at").
			Optional().
			Nillable(),
		field.Time("api_updated_at").
			Optional().
			Nillable(),
		field.String("agent_uuid").
			Optional(),
		field.String("agent_domain").
			Optional(),
		field.String("agent_version").
			Optional(),
		field.String("agent_os_type").
			Optional(),
		field.String("agent_network_status").
			Optional(),
		field.Bool("agent_infected").
			Default(false),
		field.String("agent_operational_state").
			Optional(),
	}
}

func (BronzeS1App) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id"),
		index.Fields("name"),
		index.Fields("os_type"),
		index.Fields("collected_at"),
	}
}

func (BronzeS1App) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_apps"},
	}
}
