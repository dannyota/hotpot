package inventory

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"github.com/dannyota/hotpot/pkg/schema/bronze/mixin"
)

// BronzeMEECInventoryComputer represents a managed computer in ManageEngine Endpoint Central.
type BronzeMEECInventoryComputer struct {
	ent.Schema
}

func (BronzeMEECInventoryComputer) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Timestamp{},
	}
}

func (BronzeMEECInventoryComputer) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("resource_id").
			Unique().
			Immutable().
			Comment("MEEC resource_id (integer as string)"),
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
			Optional().
			Comment("0=Unknown, 1=Windows, 2=Mac, 3=Linux"),
		field.String("os_platform_name").
			Optional(),
		field.String("os_version").
			Optional(),
		field.String("service_pack").
			Optional(),
		field.String("agent_version").
			Optional(),
		field.Int("computer_live_status").
			Optional().
			Comment("1=Live, 2=Down, 3=Unknown"),
		field.Int("installation_status").
			Optional().
			Comment("21=Yet to install, 22=Installed, 23=Uninstalled"),
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
			Optional().
			Comment("Timestamp in milliseconds"),
		field.Int64("agent_last_contact_time").
			Optional().
			Comment("Timestamp in milliseconds"),
		field.Int64("agent_installed_on").
			Optional().
			Comment("Timestamp in milliseconds"),
		field.String("customer_name").
			Optional(),
		field.Int("customer_id").
			Optional(),
	}
}

func (BronzeMEECInventoryComputer) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("resource_name"),
		index.Fields("domain_netbios_name"),
		index.Fields("os_platform"),
		index.Fields("computer_live_status"),
		index.Fields("collected_at"),
	}
}

func (BronzeMEECInventoryComputer) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "meec_inventory_computers"},
	}
}
