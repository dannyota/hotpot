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

// BronzeHistoryDOFirewall stores historical snapshots of DigitalOcean Firewalls.
type BronzeHistoryDOFirewall struct {
	ent.Schema
}

func (BronzeHistoryDOFirewall) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOFirewall) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Firewall by resource_id"),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.JSON("inbound_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("outbound_rules_json", json.RawMessage{}).
			Optional(),
		field.JSON("droplet_ids_json", json.RawMessage{}).
			Optional(),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.String("api_created_at").
			Optional(),
		field.JSON("pending_changes_json", json.RawMessage{}).
			Optional(),
	}
}

func (BronzeHistoryDOFirewall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("status"),
	}
}

func (BronzeHistoryDOFirewall) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_firewalls_history"},
	}
}
