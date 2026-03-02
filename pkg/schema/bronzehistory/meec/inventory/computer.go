package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	historymixin "github.com/dannyota/hotpot/pkg/schema/bronzehistory/mixin"
)

// BronzeHistoryMEECInventoryComputer stores historical snapshots of MEEC computers.
type BronzeHistoryMEECInventoryComputer struct {
	ent.Schema
}

func (BronzeHistoryMEECInventoryComputer) Mixin() []ent.Mixin {
	return []ent.Mixin{historymixin.Timestamp{}}
}

func (BronzeHistoryMEECInventoryComputer) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("id").StorageKey("history_id"),
		field.String("resource_id").
			NotEmpty().
			Comment("Link to bronze computer by resource_id"),

		field.String("resource_name").
			Optional(),
		field.String("fqdn_name").
			Optional(),
		field.String("domain_netbios_name").
			Optional(),
		field.String("ip_address").
			Optional(),
		field.String("mac_address").
			Optional(),
		field.String("os_name").
			Optional(),
		field.Int("os_platform").
			Optional(),
		field.String("os_platform_name").
			Optional(),
		field.String("os_version").
			Optional(),
		field.String("service_pack").
			Optional(),
		field.String("agent_version").
			Optional(),
		field.Int("computer_live_status").
			Optional(),
		field.Int("installation_status").
			Optional(),
		field.Int("managed_status").
			Optional(),
		field.String("branch_office_name").
			Optional(),
		field.String("owner").
			Optional(),
		field.String("owner_email_id").
			Optional(),
		field.String("description").
			Optional(),
		field.String("location").
			Optional(),
		field.Int64("last_sync_time").
			Optional(),
		field.Int64("agent_last_contact_time").
			Optional(),
		field.Int64("agent_installed_on").
			Optional(),
		field.String("customer_name").
			Optional(),
		field.Int("customer_id").
			Optional(),
	}
}

func (BronzeHistoryMEECInventoryComputer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_id", "valid_from"),
		index.Fields("valid_to"),
		index.Fields("collected_at"),
	}
}

func (BronzeHistoryMEECInventoryComputer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_computers_history"},
	}
}
