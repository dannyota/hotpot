package compute

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

// BronzeGreenNodeComputeServer represents a GreenNode compute server in the bronze layer.
type BronzeGreenNodeComputeServer struct {
	ent.Schema
}

func (BronzeGreenNodeComputeServer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeGreenNodeComputeServer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("Server UUID"),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("location").
			Optional(),
		field.String("zone_id").
			Optional(),
		field.String("created_at_api").
			Optional().
			Comment("Server creation timestamp from API"),
		field.String("boot_volume_id").
			Optional(),
		field.Bool("encryption_volume").
			Default(false),
		field.Bool("licence").
			Default(false),
		field.String("metadata").
			Optional(),
		field.String("migrate_state").
			Optional(),
		field.String("product").
			Optional(),
		field.String("server_group_id").
			Optional(),
		field.String("server_group_name").
			Optional(),
		field.String("ssh_key_name").
			Optional(),
		field.Bool("stop_before_migrate").
			Default(false),
		field.String("user").
			Optional(),

		// Image info (flattened)
		field.String("image_id").
			Optional(),
		field.String("image_type").
			Optional(),
		field.String("image_version").
			Optional(),

		// Flavor info (flattened)
		field.String("flavor_id").
			Optional(),
		field.String("flavor_name").
			Optional(),
		field.Int64("flavor_cpu").
			Optional(),
		field.Int64("flavor_memory").
			Optional(),
		field.Int64("flavor_gpu").
			Optional(),
		field.Int64("flavor_bandwidth").
			Optional(),

		// InterfacesJSON stores all network interfaces as JSON.
		//
		//	{"external": [...], "internal": [...]}
		field.JSON("interfaces_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("region").
			NotEmpty().
			Comment("GreenNode region (e.g. hcm-3, han-1)"),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeServer) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sec_groups", BronzeGreenNodeComputeServerSecGroup.Type),
	}
}

func (BronzeGreenNodeComputeServer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("project_id"),
		index.Fields("region"),
		index.Fields("collected_at"),
	}
}

func (BronzeGreenNodeComputeServer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_servers"},
	}
}

// BronzeGreenNodeComputeServerSecGroup represents a security group attached to a server.
type BronzeGreenNodeComputeServerSecGroup struct {
	ent.Schema
}

func (BronzeGreenNodeComputeServerSecGroup) Fields() []ent.Field {
	return []ent.Field{
		field.String("uuid").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
	}
}

func (BronzeGreenNodeComputeServerSecGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("server", BronzeGreenNodeComputeServer.Type).
			Ref("sec_groups").
			Unique().
			Required(),
	}
}

func (BronzeGreenNodeComputeServerSecGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_sec_groups"},
	}
}
