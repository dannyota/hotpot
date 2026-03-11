package s1

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryS1AppInventory stores historical snapshots of SentinelOne application inventory.
type BronzeHistoryS1AppInventory struct {
	ent.Schema
}

func (BronzeHistoryS1AppInventory) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryS1AppInventory) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze app inventory by resource_id"),

		field.String("application_name").
			NotEmpty(),
		field.String("application_vendor").
			Optional(),
		field.Int("endpoints_count").
			Optional(),
		field.Int("application_versions_count").
			Optional(),
		field.Bool("estimate").
			Default(false),
	}
}

func (BronzeHistoryS1AppInventory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryS1AppInventory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "s1_app_inventory_history"},
	}
}
