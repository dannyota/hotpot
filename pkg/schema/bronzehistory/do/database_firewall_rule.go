package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDODatabaseFirewallRule stores historical snapshots of DigitalOcean Database Firewall Rules.
type BronzeHistoryDODatabaseFirewallRule struct {
	ent.Schema
}

func (BronzeHistoryDODatabaseFirewallRule) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDODatabaseFirewallRule) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze DatabaseFirewallRule by resource_id"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("uuid").
			NotEmpty(),
		field.String("type").
			Optional(),
		field.String("value").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeHistoryDODatabaseFirewallRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("cluster_id"),
		index.Fields("type"),
	}
}

func (BronzeHistoryDODatabaseFirewallRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_firewall_rules_history"},
	}
}
