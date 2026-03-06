package loadbalancer

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "danny.vn/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeLoadBalancerCertificate stores historical snapshots of GreenNode LB certificates.
type BronzeHistoryGreenNodeLoadBalancerCertificate struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeLoadBalancerCertificate) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeLoadBalancerCertificate) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("certificate_type").
			Optional(),
		field.String("expired_at").
			Optional(),
		field.String("imported_at").
			Optional(),
		field.Int64("not_after").
			Optional(),
		field.String("key_algorithm").
			Optional(),
		field.String("serial").
			Optional(),
		field.String("subject").
			Optional(),
		field.String("domain_name").
			Optional(),
		field.Bool("in_use").
			Default(false),
		field.String("issuer").
			Optional(),
		field.String("signature_algorithm").
			Optional(),
		field.Int64("not_before").
			Optional(),
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerCertificate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeLoadBalancerCertificate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_certificates_history"},
	}
}
