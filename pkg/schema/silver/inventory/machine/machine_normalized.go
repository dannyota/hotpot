package machine

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InventoryMachineNormalized holds per-provider normalized rows before merge.
type InventoryMachineNormalized struct {
	ent.Schema
}

func (InventoryMachineNormalized) Fields() []ent.Field {
	return []ent.Field{
		// Identity — unique per provider+bronze record.
		field.String("id").StorageKey("resource_id").Unique().Immutable(),
		field.String("provider").NotEmpty(),
		field.Bool("is_base"),
		field.String("bronze_table").NotEmpty(),
		field.String("bronze_resource_id").NotEmpty(),

		// Normalized machine fields.
		field.String("hostname").Optional(),
		field.String("os_type").Optional(),
		field.String("os_name").Optional(),
		field.String("status"),
		field.String("internal_ip").Optional(),
		field.String("external_ip").Optional(),
		field.String("environment").Optional(),
		field.String("cloud_project").Optional(),
		field.String("cloud_zone").Optional(),
		field.String("cloud_machine_type").Optional(),
		field.Time("created").Optional().Nillable(),

		// Merge keys — used by merge engine for dedup.
		field.JSON("merge_keys_json", map[string][]string{}),

		// Timestamps from bronze (NOT immutable — upserts may update).
		field.Time("collected_at"),
		field.Time("first_collected_at"),

		// When this row was last normalized.
		field.Time("normalized_at"),
	}
}

func (InventoryMachineNormalized) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider"),
		index.Fields("provider", "bronze_resource_id").Unique(),
	}
}

func (InventoryMachineNormalized) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "inventory_machine_normalized"},
	}
}
