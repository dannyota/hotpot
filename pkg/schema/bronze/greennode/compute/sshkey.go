package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"danny.vn/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeComputeSSHKey represents a GreenNode SSH key in the bronze layer.
type BronzeGreenNodeComputeSSHKey struct {
	ent.Schema
}

func (BronzeGreenNodeComputeSSHKey) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeComputeSSHKey) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("SSH Key ID"),
		field.String("name").
			NotEmpty(),
		field.String("created_at_api").
			Optional().
			Comment("Key creation timestamp from API"),
		field.String("pub_key").
			Optional(),
		field.String("status").
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeSSHKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeComputeSSHKey) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_ssh_keys"},
	}
}
