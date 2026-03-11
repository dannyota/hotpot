package software

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InventorySoftwareNormalized holds per-provider normalized rows before merge.
type InventorySoftwareNormalized struct {
	ent.Schema
}

func (InventorySoftwareNormalized) Fields() []ent.Field {
	return []ent.Field{
		// Identity — unique per provider+bronze record.
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("provider").NotEmpty(),
		field.Bool("is_base"),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),

		// Normalized fields.
		field.String("machine_id").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("version").Optional(),
		field.String("publisher").Optional(),
		field.Time("installed_on").Optional().Nillable(),

		// Timestamps from bronze (NOT immutable — upserts may update).
		field.Time("collected_at"),
		field.Time("first_collected_at"),

		// When this row was last normalized.
		field.Time("normalized_at"),
	}
}

func (InventorySoftwareNormalized) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider"),
		index.Fields("provider", "bronze_resource_id").Unique(),
		index.Fields("machine_id", "name"),
	}
}

func (InventorySoftwareNormalized) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_software_normalized"},
	}
}
