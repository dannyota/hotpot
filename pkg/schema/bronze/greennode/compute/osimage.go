package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeComputeOSImage represents a GreenNode OS image in the bronze layer.
type BronzeGreenNodeComputeOSImage struct {
	ent.Schema
}

func (BronzeGreenNodeComputeOSImage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeComputeOSImage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("OS Image ID"),
		field.String("image_type").
			Optional(),
		field.String("image_version").
			Optional(),
		field.Bool("licence").
			Optional().
			Nillable(),
		field.String("license_key").
			Optional().
			Nillable(),
		field.String("description").
			Optional(),
		field.String("zone_id").
			Optional(),
		field.JSON("flavor_zone_ids", []string{}).
			Optional(),
		field.JSON("default_tag_ids", []string{}).
			Optional(),
		field.Int64("package_limit_cpu").
			Optional(),
		field.Int64("package_limit_memory").
			Optional(),
		field.Int64("package_limit_disk_size").
			Optional(),
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeOSImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeComputeOSImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_os_images"},
	}
}
