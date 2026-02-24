package loadbalancer

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGreenNodeLoadBalancerCertificate represents a GreenNode load balancer certificate in the bronze layer.
type BronzeGreenNodeLoadBalancerCertificate struct {
	ent.Schema
}

func (BronzeGreenNodeLoadBalancerCertificate) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeLoadBalancerCertificate) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Certificate UUID"),
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

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeLoadBalancerCertificate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeLoadBalancerCertificate) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_loadbalancer_certificates"},
	}
}
