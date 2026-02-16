package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDODatabaseFirewallRule represents a DigitalOcean Database Firewall Rule in the bronze layer.
type BronzeDODatabaseFirewallRule struct {
	ent.Schema
}

func (BronzeDODatabaseFirewallRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDODatabaseFirewallRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Composite ID: {clusterID}:{ruleUUID}"),
		field.String("cluster_id").
			NotEmpty(),
		field.String("uuid").
			NotEmpty().
			Comment("Firewall rule UUID"),
		field.String("type").
			Optional().
			Comment("Rule type (ip_addr, app, tag, droplet, kubernetes)"),
		field.String("value").
			Optional(),
		field.Time("api_created_at").
			Optional().
			Nillable(),
	}
}

func (BronzeDODatabaseFirewallRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("cluster_id"),
		index.Fields("type"),
		index.Fields("collected_at"),
	}
}

func (BronzeDODatabaseFirewallRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_database_firewall_rules"},
	}
}
