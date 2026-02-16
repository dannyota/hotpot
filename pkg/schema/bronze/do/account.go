package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeDOAccount represents a DigitalOcean Account in the bronze layer.
type BronzeDOAccount struct {
	ent.Schema
}

func (BronzeDOAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeDOAccount) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("DigitalOcean Account UUID"),
		field.String("email").
			Optional(),
		field.String("name").
			Optional(),
		field.String("status").
			Optional(),
		field.String("status_message").
			Optional(),
		field.Int("droplet_limit").
			Default(0),
		field.Int("floating_ip_limit").
			Default(0),
		field.Int("reserved_ip_limit").
			Default(0),
		field.Int("volume_limit").
			Default(0),
		field.Bool("email_verified").
			Default(false),
		field.String("team_name").
			Optional(),
		field.String("team_uuid").
			Optional(),
	}
}

func (BronzeDOAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("collected_at"),
	}
}

func (BronzeDOAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_accounts"},
	}
}
