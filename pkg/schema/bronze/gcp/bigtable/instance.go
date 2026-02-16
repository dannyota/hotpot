package bigtable

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeGCPBigtableInstance represents a GCP Bigtable instance in the bronze layer.
// Fields preserve raw API response data from bigtable.admin.v2.instances.list.
type BronzeGCPBigtableInstance struct {
	ent.Schema
}

func (BronzeGCPBigtableInstance) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGCPBigtableInstance) Fields() []ent.Field {
	return []ent.Field{
		// Resource name: projects/{project}/instances/{instance}
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Instance resource name"),
		field.String("display_name").
			Optional(),
		field.Int32("state").
			Optional().
			Comment("Instance state: 0=STATE_NOT_KNOWN, 1=READY, 2=CREATING"),
		field.Int32("instance_type").
			Optional().
			Comment("Instance type: 0=TYPE_UNSPECIFIED, 1=PRODUCTION, 2=DEVELOPMENT"),
		field.String("create_time").
			Optional(),
		field.Bool("satisfies_pzs").
			Optional().
			Nillable(),

		// LabelsJSON contains user-defined labels.
		//
		//	{"key1": "value1", "key2": "value2"}
		field.JSON("labels_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGCPBigtableInstance) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("clusters", BronzeGCPBigtableCluster.Type),
	}
}

func (BronzeGCPBigtableInstance) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id"),
		index.Fields("collected_at"),
	}
}

func (BronzeGCPBigtableInstance) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "gcp_bigtable_instances"},
	}
}
