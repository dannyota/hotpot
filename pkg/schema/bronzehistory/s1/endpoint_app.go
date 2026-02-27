package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1EndpointApp stores historical snapshots of endpoint applications.
type BronzeHistoryS1EndpointApp struct {
	ent.Schema
}

func (BronzeHistoryS1EndpointApp) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1EndpointApp) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze endpoint app by resource_id"),

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

func (BronzeHistoryS1EndpointApp) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("agent_id"),
	}
}

func (BronzeHistoryS1EndpointApp) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_endpoint_apps_history"},
	}
}
