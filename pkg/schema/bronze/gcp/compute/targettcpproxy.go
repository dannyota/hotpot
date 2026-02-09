package compute

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPComputeTargetTcpProxy represents a GCP Compute Engine target TCP proxy in the bronze layer.
// Fields preserve raw API response data from compute.targetTcpProxies.list.
type BronzeGCPComputeTargetTcpProxy struct {
	ent.Schema
}

func (BronzeGCPComputeTargetTcpProxy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPComputeTargetTcpProxy) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("GCP API ID, used as primary key for linking"),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.String("creation_timestamp").
			Optional(),
		field.String("self_link").
			Optional(),
		field.String("service").
			Optional().
			Comment("URL to the BackendService resource"),
		field.Bool("proxy_bind").
			Default(false),
		field.String("proxy_header").
			Optional().
			Comment("Specifies the type of proxy header to append"),
		field.String("region").
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPComputeTargetTcpProxy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPComputeTargetTcpProxy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_compute_target_tcp_proxies"},
	}
}
