-- Create "s1_accounts" table
CREATE TABLE "bronze"."s1_accounts" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "state" character varying NULL,
  "account_type" character varying NULL,
  "api_created_at" timestamptz NULL,
  "api_updated_at" timestamptz NULL,
  "expiration" timestamptz NULL,
  "unlimited_expiration" boolean NOT NULL DEFAULT false,
  "active_agents" bigint NOT NULL DEFAULT 0,
  "total_licenses" bigint NOT NULL DEFAULT 0,
  "usage_type" character varying NULL,
  "billing_mode" character varying NULL,
  "creator" character varying NULL,
  "creator_id" character varying NULL,
  "number_of_sites" bigint NOT NULL DEFAULT 0,
  "external_id" character varying NULL,
  "licenses_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1account_account_type" to table: "s1_accounts"
CREATE INDEX "bronzes1account_account_type" ON "bronze"."s1_accounts" ("account_type");
-- Create index "bronzes1account_collected_at" to table: "s1_accounts"
CREATE INDEX "bronzes1account_collected_at" ON "bronze"."s1_accounts" ("collected_at");
-- Create index "bronzes1account_state" to table: "s1_accounts"
CREATE INDEX "bronzes1account_state" ON "bronze"."s1_accounts" ("state");
-- Create "s1_agents" table
CREATE TABLE "bronze"."s1_agents" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "computer_name" character varying NOT NULL,
  "external_ip" character varying NULL,
  "site_name" character varying NULL,
  "account_id" character varying NULL,
  "account_name" character varying NULL,
  "agent_version" character varying NULL,
  "os_type" character varying NULL,
  "os_name" character varying NULL,
  "os_revision" character varying NULL,
  "os_arch" character varying NULL,
  "is_active" boolean NOT NULL DEFAULT false,
  "is_infected" boolean NOT NULL DEFAULT false,
  "is_decommissioned" boolean NOT NULL DEFAULT false,
  "machine_type" character varying NULL,
  "domain" character varying NULL,
  "uuid" character varying NULL,
  "network_status" character varying NULL,
  "last_active_date" timestamptz NULL,
  "registered_at" timestamptz NULL,
  "api_updated_at" timestamptz NULL,
  "os_start_time" timestamptz NULL,
  "active_threats" bigint NOT NULL DEFAULT 0,
  "encrypted_applications" boolean NOT NULL DEFAULT false,
  "group_name" character varying NULL,
  "group_id" character varying NULL,
  "cpu_count" bigint NOT NULL DEFAULT 0,
  "core_count" bigint NOT NULL DEFAULT 0,
  "cpu_id" character varying NULL,
  "total_memory" bigint NOT NULL DEFAULT 0,
  "model_name" character varying NULL,
  "serial_number" character varying NULL,
  "storage_encryption_status" character varying NULL,
  "network_interfaces_json" jsonb NULL,
  "site_id" character varying NULL,
  "api_created_at" timestamptz NULL,
  "os_username" character varying NULL,
  "group_ip" character varying NULL,
  "scan_status" character varying NULL,
  "scan_started_at" timestamptz NULL,
  "scan_finished_at" timestamptz NULL,
  "mitigation_mode" character varying NULL,
  "mitigation_mode_suspicious" character varying NULL,
  "last_logged_in_user_name" character varying NULL,
  "installer_type" character varying NULL,
  "external_id" character varying NULL,
  "last_ip_to_mgmt" character varying NULL,
  "is_up_to_date" boolean NOT NULL DEFAULT false,
  "is_pending_uninstall" boolean NOT NULL DEFAULT false,
  "is_uninstalled" boolean NOT NULL DEFAULT false,
  "apps_vulnerability_status" character varying NULL,
  "console_migration_status" character varying NULL,
  "ranger_version" character varying NULL,
  "ranger_status" character varying NULL,
  "active_directory_json" jsonb NULL,
  "locations_json" jsonb NULL,
  "user_actions_needed_json" jsonb NULL,
  "missing_permissions_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1agent_account_id" to table: "s1_agents"
CREATE INDEX "bronzes1agent_account_id" ON "bronze"."s1_agents" ("account_id");
-- Create index "bronzes1agent_collected_at" to table: "s1_agents"
CREATE INDEX "bronzes1agent_collected_at" ON "bronze"."s1_agents" ("collected_at");
-- Create index "bronzes1agent_is_active" to table: "s1_agents"
CREATE INDEX "bronzes1agent_is_active" ON "bronze"."s1_agents" ("is_active");
-- Create index "bronzes1agent_is_infected" to table: "s1_agents"
CREATE INDEX "bronzes1agent_is_infected" ON "bronze"."s1_agents" ("is_infected");
-- Create index "bronzes1agent_os_type" to table: "s1_agents"
CREATE INDEX "bronzes1agent_os_type" ON "bronze"."s1_agents" ("os_type");
-- Create index "bronzes1agent_scan_status" to table: "s1_agents"
CREATE INDEX "bronzes1agent_scan_status" ON "bronze"."s1_agents" ("scan_status");
-- Create index "bronzes1agent_site_id" to table: "s1_agents"
CREATE INDEX "bronzes1agent_site_id" ON "bronze"."s1_agents" ("site_id");
-- Create "s1_groups" table
CREATE TABLE "bronze"."s1_groups" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "site_id" character varying NULL,
  "type" character varying NULL,
  "is_default" boolean NOT NULL DEFAULT false,
  "inherits" boolean NOT NULL DEFAULT false,
  "rank" bigint NULL,
  "total_agents" bigint NOT NULL DEFAULT 0,
  "creator" character varying NULL,
  "creator_id" character varying NULL,
  "filter_name" character varying NULL,
  "filter_id" character varying NULL,
  "api_created_at" timestamptz NULL,
  "api_updated_at" timestamptz NULL,
  "registration_token" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1group_collected_at" to table: "s1_groups"
CREATE INDEX "bronzes1group_collected_at" ON "bronze"."s1_groups" ("collected_at");
-- Create index "bronzes1group_site_id" to table: "s1_groups"
CREATE INDEX "bronzes1group_site_id" ON "bronze"."s1_groups" ("site_id");
-- Create index "bronzes1group_type" to table: "s1_groups"
CREATE INDEX "bronzes1group_type" ON "bronze"."s1_groups" ("type");
-- Create "s1_ranger_devices" table
CREATE TABLE "bronze"."s1_ranger_devices" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "local_ip" character varying NULL,
  "external_ip" character varying NULL,
  "mac_address" character varying NULL,
  "os_type" character varying NULL,
  "os_name" character varying NULL,
  "os_version" character varying NULL,
  "device_type" character varying NULL,
  "device_function" character varying NULL,
  "manufacturer" character varying NULL,
  "managed_state" character varying NULL,
  "agent_id" character varying NULL,
  "first_seen" timestamptz NULL,
  "last_seen" timestamptz NULL,
  "subnet_address" character varying NULL,
  "gateway_ip_address" character varying NULL,
  "gateway_mac_address" character varying NULL,
  "network_name" character varying NULL,
  "domain" character varying NULL,
  "site_name" character varying NULL,
  "device_review" character varying NULL,
  "has_identity" boolean NOT NULL DEFAULT false,
  "has_user_label" boolean NOT NULL DEFAULT false,
  "fingerprint_score" bigint NULL,
  "tcp_ports_json" jsonb NULL,
  "udp_ports_json" jsonb NULL,
  "hostnames_json" jsonb NULL,
  "discovery_methods_json" jsonb NULL,
  "networks_json" jsonb NULL,
  "tags_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1rangerdevice_agent_id" to table: "s1_ranger_devices"
CREATE INDEX "bronzes1rangerdevice_agent_id" ON "bronze"."s1_ranger_devices" ("agent_id");
-- Create index "bronzes1rangerdevice_collected_at" to table: "s1_ranger_devices"
CREATE INDEX "bronzes1rangerdevice_collected_at" ON "bronze"."s1_ranger_devices" ("collected_at");
-- Create index "bronzes1rangerdevice_managed_state" to table: "s1_ranger_devices"
CREATE INDEX "bronzes1rangerdevice_managed_state" ON "bronze"."s1_ranger_devices" ("managed_state");
-- Create index "bronzes1rangerdevice_network_name" to table: "s1_ranger_devices"
CREATE INDEX "bronzes1rangerdevice_network_name" ON "bronze"."s1_ranger_devices" ("network_name");
-- Create index "bronzes1rangerdevice_site_name" to table: "s1_ranger_devices"
CREATE INDEX "bronzes1rangerdevice_site_name" ON "bronze"."s1_ranger_devices" ("site_name");
-- Create "s1_ranger_gateways" table
CREATE TABLE "bronze"."s1_ranger_gateways" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "ip" character varying NULL,
  "mac_address" character varying NULL,
  "external_ip" character varying NULL,
  "manufacturer" character varying NULL,
  "network_name" character varying NULL,
  "account_id" character varying NULL,
  "account_name" character varying NULL,
  "site_id" character varying NULL,
  "number_of_agents" bigint NULL,
  "number_of_rangers" bigint NULL,
  "connected_rangers" bigint NULL,
  "total_agents" bigint NULL,
  "agent_percentage" double precision NULL,
  "allow_scan" boolean NOT NULL DEFAULT false,
  "archived" boolean NOT NULL DEFAULT false,
  "new_network" boolean NOT NULL DEFAULT false,
  "inherit_settings" boolean NOT NULL DEFAULT false,
  "tcp_port_scan" boolean NOT NULL DEFAULT false,
  "udp_port_scan" boolean NOT NULL DEFAULT false,
  "icmp_scan" boolean NOT NULL DEFAULT false,
  "smb_scan" boolean NOT NULL DEFAULT false,
  "mdns_scan" boolean NOT NULL DEFAULT false,
  "rdns_scan" boolean NOT NULL DEFAULT false,
  "snmp_scan" boolean NOT NULL DEFAULT false,
  "scan_only_local_subnets" boolean NOT NULL DEFAULT false,
  "created_at_api" timestamptz NULL,
  "expiry_date" timestamptz NULL,
  "tcp_ports_json" jsonb NULL,
  "udp_ports_json" jsonb NULL,
  "restrictions_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1rangergateway_account_id" to table: "s1_ranger_gateways"
CREATE INDEX "bronzes1rangergateway_account_id" ON "bronze"."s1_ranger_gateways" ("account_id");
-- Create index "bronzes1rangergateway_collected_at" to table: "s1_ranger_gateways"
CREATE INDEX "bronzes1rangergateway_collected_at" ON "bronze"."s1_ranger_gateways" ("collected_at");
-- Create index "bronzes1rangergateway_network_name" to table: "s1_ranger_gateways"
CREATE INDEX "bronzes1rangergateway_network_name" ON "bronze"."s1_ranger_gateways" ("network_name");
-- Create index "bronzes1rangergateway_site_id" to table: "s1_ranger_gateways"
CREATE INDEX "bronzes1rangergateway_site_id" ON "bronze"."s1_ranger_gateways" ("site_id");
-- Create "s1_ranger_settings" table
CREATE TABLE "bronze"."s1_ranger_settings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "account_id" character varying NULL,
  "scope_id" character varying NULL,
  "enabled" boolean NOT NULL DEFAULT false,
  "use_periodic_snapshots" boolean NOT NULL DEFAULT false,
  "snapshot_period" bigint NULL,
  "network_decommission_value" bigint NULL,
  "min_agents_in_network_to_scan" bigint NULL,
  "tcp_port_scan" boolean NOT NULL DEFAULT false,
  "udp_port_scan" boolean NOT NULL DEFAULT false,
  "icmp_scan" boolean NOT NULL DEFAULT false,
  "smb_scan" boolean NOT NULL DEFAULT false,
  "mdns_scan" boolean NOT NULL DEFAULT false,
  "rdns_scan" boolean NOT NULL DEFAULT false,
  "snmp_scan" boolean NOT NULL DEFAULT false,
  "multi_scan_ssdp" boolean NOT NULL DEFAULT false,
  "use_full_dns_scan" boolean NOT NULL DEFAULT false,
  "scan_only_local_subnets" boolean NOT NULL DEFAULT false,
  "auto_enable_networks" boolean NOT NULL DEFAULT false,
  "combine_devices" boolean NOT NULL DEFAULT false,
  "new_network_in_hours" bigint NULL,
  "tcp_ports_json" jsonb NULL,
  "udp_ports_json" jsonb NULL,
  "restrictions_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1rangersetting_account_id" to table: "s1_ranger_settings"
CREATE INDEX "bronzes1rangersetting_account_id" ON "bronze"."s1_ranger_settings" ("account_id");
-- Create index "bronzes1rangersetting_collected_at" to table: "s1_ranger_settings"
CREATE INDEX "bronzes1rangersetting_collected_at" ON "bronze"."s1_ranger_settings" ("collected_at");
-- Create "s1_sites" table
CREATE TABLE "bronze"."s1_sites" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "account_id" character varying NULL,
  "account_name" character varying NULL,
  "state" character varying NULL,
  "site_type" character varying NULL,
  "suite" character varying NULL,
  "creator" character varying NULL,
  "creator_id" character varying NULL,
  "health_status" boolean NOT NULL DEFAULT false,
  "active_licenses" bigint NOT NULL DEFAULT 0,
  "total_licenses" bigint NOT NULL DEFAULT 0,
  "unlimited_licenses" boolean NOT NULL DEFAULT false,
  "is_default" boolean NOT NULL DEFAULT false,
  "description" character varying NULL,
  "api_created_at" timestamptz NULL,
  "expiration" timestamptz NULL,
  "api_updated_at" timestamptz NULL,
  "external_id" character varying NULL,
  "sku" character varying NULL,
  "usage_type" character varying NULL,
  "unlimited_expiration" boolean NOT NULL DEFAULT false,
  "inherit_account_expiration" boolean NOT NULL DEFAULT false,
  "licenses_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzes1site_account_id" to table: "s1_sites"
CREATE INDEX "bronzes1site_account_id" ON "bronze"."s1_sites" ("account_id");
-- Create index "bronzes1site_collected_at" to table: "s1_sites"
CREATE INDEX "bronzes1site_collected_at" ON "bronze"."s1_sites" ("collected_at");
-- Create index "bronzes1site_state" to table: "s1_sites"
CREATE INDEX "bronzes1site_state" ON "bronze"."s1_sites" ("state");
-- Create "s1_agent_nics" table
CREATE TABLE "bronze"."s1_agent_nics" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "interface_id" character varying NULL,
  "name" character varying NULL,
  "description" character varying NULL,
  "type" character varying NULL,
  "inet_json" jsonb NULL,
  "inet6_json" jsonb NULL,
  "physical" character varying NULL,
  "gateway_ip" character varying NULL,
  "gateway_mac" character varying NULL,
  "bronze_s1agent_nics" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "s1_agent_nics_s1_agents_nics" FOREIGN KEY ("bronze_s1agent_nics") REFERENCES "bronze"."s1_agents" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
