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

// BronzeDOFirewall represents a DigitalOcean Cloud Firewall in the bronze layer.
type BronzeDOFirewall struct {
	ent.Schema
}

func (BronzeDOFirewall) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOFirewall) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Firewall UUID"),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.JSON("inbound_rules_json", json.RawMessage{}).
			Optional().
			Comment("Raw inbound rules JSON"),
		field.JSON("outbound_rules_json", json.RawMessage{}).
			Optional().
			Comment("Raw outbound rules JSON"),
		field.JSON("droplet_ids_json", json.RawMessage{}).
			Optional().
			Comment("Associated droplet IDs"),
		field.JSON("tags_json", json.RawMessage{}).
			Optional(),
		field.String("api_created_at").
			Optional().
			Comment("API-reported creation timestamp"),
		field.JSON("pending_changes_json", json.RawMessage{}).
			Optional().
			Comment("Pending firewall changes"),
	}
}

func (BronzeDOFirewall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOFirewall) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_firewalls"},
	}
}
