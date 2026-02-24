package glb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeGLBGlobalPackage represents a GreenNode global load balancer package in the bronze layer.
type BronzeGreenNodeGLBGlobalPackage struct {
	ent.Schema
}

func (BronzeGreenNodeGLBGlobalPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeGLBGlobalPackage) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Global package ID"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("description_en").
			Optional(),
		field.JSON("detail_json", json.RawMessage{}).
			Optional().
			Comment("Detail as JSONB"),
		field.Bool("enabled").
			Default(false),
		field.String("base_sku").
			Optional(),
		field.Int("base_connection_rate").
			Optional(),
		field.Int("base_domestic_traffic_total").
			Optional(),
		field.Int("base_non_domestic_traffic_total").
			Optional(),
		field.String("connection_sku").
			Optional(),
		field.String("domestic_traffic_sku").
			Optional(),
		field.String("non_domestic_traffic_sku").
			Optional(),
		field.String("created_at_api").
			Optional(),
		field.String("updated_at_api").
			Optional(),
		field.JSON("vlb_packages_json", json.RawMessage{}).
			Optional().
			Comment("VLB packages as JSONB"),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeGLBGlobalPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeGLBGlobalPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_packages"},
	}
}
