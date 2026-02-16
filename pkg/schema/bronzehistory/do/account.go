package do

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryDOAccount stores historical snapshots of DigitalOcean Accounts.
type BronzeHistoryDOAccount struct {
	ent.Schema
}

func (BronzeHistoryDOAccount) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryDOAccount) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").
			Unique().
			Immutable(),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze Account by resource_id"),
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

func (BronzeHistoryDOAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("status"),
	}
}

func (BronzeHistoryDOAccount) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "do_accounts_history"},
	}
}
