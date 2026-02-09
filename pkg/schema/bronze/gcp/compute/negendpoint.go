package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

type BronzeGCPComputeNegEndpoint struct {
	ent.Schema
}

func (BronzeGCPComputeNegEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{mixin.Timestamp{}}
}

func (BronzeGCPComputeNegEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Synthetic ID: {neg_resource_id}/{ip_address}:{port}"),
		field.String("instance").Optional(),
		field.String("ip_address").Optional(),
		field.String("ipv6_address").Optional(),
		field.String("port").Optional(),
		field.String("fqdn").Optional(),
		field.JSON("annotations_json", map[string]interface{}{}).Optional(),
		field.String("neg_name").NotEmpty().Comment("Parent NEG name"),
		field.String("zone").Optional(),
		field.String("project_id").NotEmpty(),
	}
}

func (BronzeGCPComputeNegEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("neg_name"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeNegEndpoint) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_neg_endpoints"},
	}
}
