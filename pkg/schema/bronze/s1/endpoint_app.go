package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeS1EndpointApp represents an application installed on a SentinelOne endpoint.
type BronzeS1EndpointApp struct {
	ent.Schema
}

func (BronzeS1EndpointApp) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeS1EndpointApp) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Synthesized: agentId||name||version"),
		field.String("agent_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("version").
			Optional(),
		field.String("publisher").
			Optional(),
		field.Int("size").
			Optional(),
		field.Time("installed_date").
			Optional().
			Nillable(),
	}
}

func (BronzeS1EndpointApp) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("agent_id"),
		index.Fields("name"),
		index.Fields("collected_at"),
		index.Fields("name", "version"),
	}
}

func (BronzeS1EndpointApp) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_endpoint_apps"},
	}
}
