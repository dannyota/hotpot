package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "hotpot/pkg/schema/bronzehistory/mixin"
)

type BronzeHistoryGCPComputeNegEndpoint struct {
	ent.Schema
}

func (BronzeHistoryGCPComputeNegEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGCPComputeNegEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("history_id").Unique().Immutable(),
		field.String("resource_id").NotEmpty().Comment("Link to bronze NEG endpoint"),
		field.String("instance").Optional(),
		field.String("ip_address").Optional(),
		field.String("ipv6_address").Optional(),
		field.String("port").Optional(),
		field.String("fqdn").Optional(),
		field.JSON("annotations_json", map[string]interface{}{}).Optional(),
		field.String("neg_name").NotEmpty(),
		field.String("zone").Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeHistoryGCPComputeNegEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryGCPComputeNegEndpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_neg_endpoints_history"},
	}
}
