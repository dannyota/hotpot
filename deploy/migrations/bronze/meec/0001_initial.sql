-- Add new schema named "bronze"
CREATE SCHEMA IF NOT EXISTS "bronze";
-- Create "meec_inventory_computers" table
CREATE TABLE "bronze"."meec_inventory_computers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_name" character varying NULL,
  "fqdn_name" character varying NULL,
  "domain_netbios_name" character varying NULL,
  "ip_address" character varying NULL,
  "mac_address" character varying NULL,
  "os_name" character varying NULL,
  "os_platform" bigint NULL,
  "os_platform_name" character varying NULL,
  "os_version" character varying NULL,
  "service_pack" character varying NULL,
  "agent_version" character varying NULL,
  "computer_live_status" bigint NULL,
  "installation_status" bigint NULL,
  "managed_status" bigint NULL,
  "branch_office_name" character varying NULL,
  "owner" character varying NULL,
  "owner_email_id" character varying NULL,
  "description" character varying NULL,
  "location" character varying NULL,
  "last_sync_time" bigint NULL,
  "agent_last_contact_time" bigint NULL,
  "agent_installed_on" bigint NULL,
  "customer_name" character varying NULL,
  "customer_id" bigint NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzemeecinventorycomputer_collected_at" to table: "meec_inventory_computers"
CREATE INDEX "bronzemeecinventorycomputer_collected_at" ON "bronze"."meec_inventory_computers" ("collected_at");
-- Create index "bronzemeecinventorycomputer_computer_live_status" to table: "meec_inventory_computers"
CREATE INDEX "bronzemeecinventorycomputer_computer_live_status" ON "bronze"."meec_inventory_computers" ("computer_live_status");
-- Create index "bronzemeecinventorycomputer_domain_netbios_name" to table: "meec_inventory_computers"
CREATE INDEX "bronzemeecinventorycomputer_domain_netbios_name" ON "bronze"."meec_inventory_computers" ("domain_netbios_name");
-- Create index "bronzemeecinventorycomputer_os_platform" to table: "meec_inventory_computers"
CREATE INDEX "bronzemeecinventorycomputer_os_platform" ON "bronze"."meec_inventory_computers" ("os_platform");
-- Create index "bronzemeecinventorycomputer_resource_name" to table: "meec_inventory_computers"
CREATE INDEX "bronzemeecinventorycomputer_resource_name" ON "bronze"."meec_inventory_computers" ("resource_name");
-- Create "meec_inventory_installed_software" table
CREATE TABLE "bronze"."meec_inventory_installed_software" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "computer_resource_id" character varying NOT NULL,
  "software_id" bigint NOT NULL,
  "software_name" character varying NULL,
  "software_version" character varying NULL,
  "display_name" character varying NULL,
  "manufacturer_name" character varying NULL,
  "installed_date" bigint NULL,
  "architecture" character varying NULL,
  "location" character varying NULL,
  "sw_type" bigint NULL,
  "sw_category_name" character varying NULL,
  "detected_time" bigint NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzemeecinventoryinstalledsoftware_collected_at" to table: "meec_inventory_installed_software"
CREATE INDEX "bronzemeecinventoryinstalledsoftware_collected_at" ON "bronze"."meec_inventory_installed_software" ("collected_at");
-- Create index "bronzemeecinventoryinstalledsoftware_computer_resource_id" to table: "meec_inventory_installed_software"
CREATE INDEX "bronzemeecinventoryinstalledsoftware_computer_resource_id" ON "bronze"."meec_inventory_installed_software" ("computer_resource_id");
-- Create index "bronzemeecinventoryinstalledsoftware_software_name" to table: "meec_inventory_installed_software"
CREATE INDEX "bronzemeecinventoryinstalledsoftware_software_name" ON "bronze"."meec_inventory_installed_software" ("software_name");
-- Create "meec_inventory_software" table
CREATE TABLE "bronze"."meec_inventory_software" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "software_name" character varying NOT NULL,
  "software_version" character varying NULL,
  "display_name" character varying NULL,
  "manufacturer_id" bigint NULL,
  "manufacturer_name" character varying NULL,
  "sw_category_name" character varying NULL,
  "sw_type" bigint NULL,
  "sw_family" bigint NULL,
  "installed_format" character varying NULL,
  "is_usage_prohibited" bigint NULL,
  "managed_installations" bigint NULL,
  "network_installations" bigint NULL,
  "managed_sw_id" bigint NULL,
  "detected_time" bigint NULL,
  "compliant_status" character varying NULL,
  "total_copies" character varying NULL,
  "remaining_copies" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzemeecinventorysoftware_collected_at" to table: "meec_inventory_software"
CREATE INDEX "bronzemeecinventorysoftware_collected_at" ON "bronze"."meec_inventory_software" ("collected_at");
-- Create index "bronzemeecinventorysoftware_manufacturer_name" to table: "meec_inventory_software"
CREATE INDEX "bronzemeecinventorysoftware_manufacturer_name" ON "bronze"."meec_inventory_software" ("manufacturer_name");
-- Create index "bronzemeecinventorysoftware_software_name" to table: "meec_inventory_software"
CREATE INDEX "bronzemeecinventorysoftware_software_name" ON "bronze"."meec_inventory_software" ("software_name");
