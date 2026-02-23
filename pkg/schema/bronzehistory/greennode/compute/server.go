package compute

import (
	"encoding/json"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryGreenNodeComputeServer stores historical snapshots of GreenNode servers.
type BronzeHistoryGreenNodeComputeServer struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeServer) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryGreenNodeComputeServer) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
		field.String("status").
			Optional(),
		field.String("location").
			Optional(),
		field.String("zone_id").
			Optional(),
		field.String("created_at_api").
			Optional(),
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

		// InterfacesJSON
		field.JSON("interfaces_json", json.RawMessage{}).
			Optional(),

		// Collection metadata
		field.String("region").
			NotEmpty(),
		field.String("project_id").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeServer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
		index.Fields("project_id"),
	}
}

func (BronzeHistoryGreenNodeComputeServer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_servers_history"},
	}
}

// BronzeHistoryGreenNodeComputeServerSecGroup stores historical snapshots of server sec groups.
type BronzeHistoryGreenNodeComputeServerSecGroup struct {
	ent.Schema
}

func (BronzeHistoryGreenNodeComputeServerSecGroup) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.Uint("server_history_id").
			Comment("Links to parent BronzeHistoryGreenNodeComputeServer"),
		field.Time("valid_from").
			Immutable(),
		field.Time("valid_to").
			Optional().
			Nillable(),
		field.String("uuid").
			NotEmpty(),
		field.String("name").
			NotEmpty(),
	}
}

func (BronzeHistoryGreenNodeComputeServerSecGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("server_history_id"),
		index.Fields("valid_from"),
		index.Fields("valid_to"),
	}
}

func (BronzeHistoryGreenNodeComputeServerSecGroup) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "greennode_compute_server_sec_groups_history"},
	}
}
