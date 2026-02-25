package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeComputeOSImage stores historical snapshots of GreenNode OS images.
type BronzeHistoryGreenNodeComputeOSImage struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeOSImage) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeComputeOSImage) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
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
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeOSImage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeComputeOSImage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_os_images_history"},
	}
}
