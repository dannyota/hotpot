-- Create "s1_app_inventory" table
CREATE TABLE "bronze"."s1_app_inventory" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "application_name" character varying NOT NULL,
  "application_vendor" character varying NULL,
  "endpoints_count" bigint NULL,
  "application_versions_count" bigint NULL,
  "estimate" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1appinventory_application_name" to table: "s1_app_inventory"
CREATE INDEX "bronzes1appinventory_application_name" ON "bronze"."s1_app_inventory" ("application_name");
-- Create index "bronzes1appinventory_application_vendor" to table: "s1_app_inventory"
CREATE INDEX "bronzes1appinventory_application_vendor" ON "bronze"."s1_app_inventory" ("application_vendor");
-- Create index "bronzes1appinventory_collected_at" to table: "s1_app_inventory"
CREATE INDEX "bronzes1appinventory_collected_at" ON "bronze"."s1_app_inventory" ("collected_at");
-- Create "s1_endpoint_apps" table
CREATE TABLE "bronze"."s1_endpoint_apps" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "agent_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "version" character varying NULL,
  "publisher" character varying NULL,
  "size" bigint NULL,
  "installed_date" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1endpointapp_agent_id" to table: "s1_endpoint_apps"
CREATE INDEX "bronzes1endpointapp_agent_id" ON "bronze"."s1_endpoint_apps" ("agent_id");
-- Create index "bronzes1endpointapp_collected_at" to table: "s1_endpoint_apps"
CREATE INDEX "bronzes1endpointapp_collected_at" ON "bronze"."s1_endpoint_apps" ("collected_at");
-- Create index "bronzes1endpointapp_name" to table: "s1_endpoint_apps"
CREATE INDEX "bronzes1endpointapp_name" ON "bronze"."s1_endpoint_apps" ("name");
-- Create index "bronzes1endpointapp_name_version" to table: "s1_endpoint_apps"
CREATE INDEX "bronzes1endpointapp_name_version" ON "bronze"."s1_endpoint_apps" ("name", "version");
-- Create "s1_network_discoveries" table
CREATE TABLE "bronze"."s1_network_discoveries" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NULL,
  "ip_address" character varying NULL,
  "domain" character varying NULL,
  "serial_number" character varying NULL,
  "category" character varying NULL,
  "sub_category" character varying NULL,
  "resource_type" character varying NULL,
  "os" character varying NULL,
  "os_family" character varying NULL,
  "os_version" character varying NULL,
  "os_name_version" character varying NULL,
  "architecture" character varying NULL,
  "manufacturer" character varying NULL,
  "cpu" character varying NULL,
  "memory_readable" character varying NULL,
  "network_name" character varying NULL,
  "asset_status" character varying NULL,
  "asset_criticality" character varying NULL,
  "asset_environment" character varying NULL,
  "infection_status" character varying NULL,
  "device_review" character varying NULL,
  "epp_unsupported_unknown" character varying NULL,
  "asset_contact_email" character varying NULL,
  "legacy_identity_policy_name" character varying NULL,
  "previous_os_type" character varying NULL,
  "previous_os_version" character varying NULL,
  "previous_device_function" character varying NULL,
  "detected_from_site" character varying NULL,
  "s1_account_id" character varying NULL,
  "s1_account_name" character varying NULL,
  "s1_site_id" character varying NULL,
  "s1_site_name" character varying NULL,
  "s1_group_id" character varying NULL,
  "s1_group_name" character varying NULL,
  "s1_scope_id" character varying NULL,
  "s1_scope_level" character varying NULL,
  "s1_scope_path" character varying NULL,
  "s1_onboarded_account_name" character varying NULL,
  "s1_onboarded_group_name" character varying NULL,
  "s1_onboarded_site_name" character varying NULL,
  "s1_onboarded_scope_level" character varying NULL,
  "s1_onboarded_scope_path" character varying NULL,
  "memory" bigint NULL,
  "core_count" bigint NULL,
  "s1_management_id" bigint NULL,
  "s1_scope_type" bigint NULL,
  "s1_onboarded_account_id" bigint NULL,
  "s1_onboarded_group_id" bigint NULL,
  "s1_onboarded_scope_id" bigint NULL,
  "s1_onboarded_site_id" bigint NULL,
  "is_ad_connector" boolean NOT NULL DEFAULT false,
  "is_dc_server" boolean NOT NULL DEFAULT false,
  "ads_enabled" boolean NOT NULL DEFAULT false,
  "first_seen_dt" timestamptz NULL,
  "last_update_dt" timestamptz NULL,
  "last_active_dt" timestamptz NULL,
  "last_reboot_dt" timestamptz NULL,
  "s1_updated_at" timestamptz NULL,
  "agent_json" jsonb NULL,
  "network_interfaces_json" jsonb NULL,
  "alerts_json" jsonb NULL,
  "alerts_count_json" jsonb NULL,
  "device_review_log_json" jsonb NULL,
  "identity_json" jsonb NULL,
  "notes_json" jsonb NULL,
  "tags_json" jsonb NULL,
  "missing_coverage_json" jsonb NULL,
  "subnets_json" jsonb NULL,
  "surfaces_json" jsonb NULL,
  "network_names_json" jsonb NULL,
  "risk_factors_json" jsonb NULL,
  "active_coverage_json" jsonb NULL,
  "discovery_methods_json" jsonb NULL,
  "hostnames_json" jsonb NULL,
  "internal_ips_json" jsonb NULL,
  "internal_ips_v6_json" jsonb NULL,
  "mac_addresses_json" jsonb NULL,
  "gateway_ips_json" jsonb NULL,
  "gateway_macs_json" jsonb NULL,
  "tcp_ports_json" jsonb NULL,
  "udp_ports_json" jsonb NULL,
  "ranger_tags_json" jsonb NULL,
  "id_secondary_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1networkdiscovery_asset_status" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_asset_status" ON "bronze"."s1_network_discoveries" ("asset_status");
-- Create index "bronzes1networkdiscovery_collected_at" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_collected_at" ON "bronze"."s1_network_discoveries" ("collected_at");
-- Create index "bronzes1networkdiscovery_device_review" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_device_review" ON "bronze"."s1_network_discoveries" ("device_review");
-- Create index "bronzes1networkdiscovery_infection_status" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_infection_status" ON "bronze"."s1_network_discoveries" ("infection_status");
-- Create index "bronzes1networkdiscovery_s1_account_id" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_s1_account_id" ON "bronze"."s1_network_discoveries" ("s1_account_id");
-- Create index "bronzes1networkdiscovery_s1_site_id" to table: "s1_network_discoveries"
CREATE INDEX "bronzes1networkdiscovery_s1_site_id" ON "bronze"."s1_network_discoveries" ("s1_site_id");
