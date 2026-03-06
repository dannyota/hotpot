package glb

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeGLBGlobalPackage stores historical snapshots of GreenNode global LB packages.
type BronzeHistoryGreenNodeGLBGlobalPackage struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeGLBGlobalPackage) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeGLBGlobalPackage) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("description_en").
			Optional(),
		field.JSON("detail_json", json.RawMessage{}).
			Optional(),
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
			Optional(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeGLBGlobalPackage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_glb_global_packages_history"},
	}
}
