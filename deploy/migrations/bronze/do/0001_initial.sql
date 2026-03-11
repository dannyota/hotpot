-- Add new schema named "bronze"
CREATE SCHEMA IF NOT EXISTS "bronze";
-- Create "do_accounts" table
CREATE TABLE "bronze"."do_accounts" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "email" character varying NULL,
  "name" character varying NULL,
  "status" character varying NULL,
  "status_message" character varying NULL,
  "droplet_limit" bigint NOT NULL DEFAULT 0,
  "floating_ip_limit" bigint NOT NULL DEFAULT 0,
  "reserved_ip_limit" bigint NOT NULL DEFAULT 0,
  "volume_limit" bigint NOT NULL DEFAULT 0,
  "email_verified" boolean NOT NULL DEFAULT false,
  "team_name" character varying NULL,
  "team_uuid" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedoaccount_collected_at" to table: "do_accounts"
CREATE INDEX "bronzedoaccount_collected_at" ON "bronze"."do_accounts" ("collected_at");
-- Create index "bronzedoaccount_status" to table: "do_accounts"
CREATE INDEX "bronzedoaccount_status" ON "bronze"."do_accounts" ("status");
-- Create "do_database_backups" table
CREATE TABLE "bronze"."do_database_backups" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "size_gigabytes" double precision NOT NULL DEFAULT 0,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabasebackup_cluster_id" to table: "do_database_backups"
CREATE INDEX "bronzedodatabasebackup_cluster_id" ON "bronze"."do_database_backups" ("cluster_id");
-- Create index "bronzedodatabasebackup_collected_at" to table: "do_database_backups"
CREATE INDEX "bronzedodatabasebackup_collected_at" ON "bronze"."do_database_backups" ("collected_at");
-- Create "do_database_configs" table
CREATE TABLE "bronze"."do_database_configs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "engine_slug" character varying NULL,
  "config_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabaseconfig_cluster_id" to table: "do_database_configs"
CREATE INDEX "bronzedodatabaseconfig_cluster_id" ON "bronze"."do_database_configs" ("cluster_id");
-- Create index "bronzedodatabaseconfig_collected_at" to table: "do_database_configs"
CREATE INDEX "bronzedodatabaseconfig_collected_at" ON "bronze"."do_database_configs" ("collected_at");
-- Create index "bronzedodatabaseconfig_engine_slug" to table: "do_database_configs"
CREATE INDEX "bronzedodatabaseconfig_engine_slug" ON "bronze"."do_database_configs" ("engine_slug");
-- Create "do_database_firewall_rules" table
CREATE TABLE "bronze"."do_database_firewall_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "uuid" character varying NOT NULL,
  "type" character varying NULL,
  "value" character varying NULL,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabasefirewallrule_cluster_id" to table: "do_database_firewall_rules"
CREATE INDEX "bronzedodatabasefirewallrule_cluster_id" ON "bronze"."do_database_firewall_rules" ("cluster_id");
-- Create index "bronzedodatabasefirewallrule_collected_at" to table: "do_database_firewall_rules"
CREATE INDEX "bronzedodatabasefirewallrule_collected_at" ON "bronze"."do_database_firewall_rules" ("collected_at");
-- Create index "bronzedodatabasefirewallrule_type" to table: "do_database_firewall_rules"
CREATE INDEX "bronzedodatabasefirewallrule_type" ON "bronze"."do_database_firewall_rules" ("type");
-- Create "do_database_pools" table
CREATE TABLE "bronze"."do_database_pools" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "user" character varying NULL,
  "size" bigint NOT NULL DEFAULT 0,
  "database" character varying NULL,
  "mode" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabasepool_cluster_id" to table: "do_database_pools"
CREATE INDEX "bronzedodatabasepool_cluster_id" ON "bronze"."do_database_pools" ("cluster_id");
-- Create index "bronzedodatabasepool_collected_at" to table: "do_database_pools"
CREATE INDEX "bronzedodatabasepool_collected_at" ON "bronze"."do_database_pools" ("collected_at");
-- Create index "bronzedodatabasepool_mode" to table: "do_database_pools"
CREATE INDEX "bronzedodatabasepool_mode" ON "bronze"."do_database_pools" ("mode");
-- Create "do_database_replicas" table
CREATE TABLE "bronze"."do_database_replicas" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "region" character varying NULL,
  "status" character varying NULL,
  "size" character varying NULL,
  "storage_size_mib" bigint NOT NULL DEFAULT 0,
  "private_network_uuid" character varying NULL,
  "tags_json" jsonb NULL,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabasereplica_cluster_id" to table: "do_database_replicas"
CREATE INDEX "bronzedodatabasereplica_cluster_id" ON "bronze"."do_database_replicas" ("cluster_id");
-- Create index "bronzedodatabasereplica_collected_at" to table: "do_database_replicas"
CREATE INDEX "bronzedodatabasereplica_collected_at" ON "bronze"."do_database_replicas" ("collected_at");
-- Create index "bronzedodatabasereplica_region" to table: "do_database_replicas"
CREATE INDEX "bronzedodatabasereplica_region" ON "bronze"."do_database_replicas" ("region");
-- Create index "bronzedodatabasereplica_status" to table: "do_database_replicas"
CREATE INDEX "bronzedodatabasereplica_status" ON "bronze"."do_database_replicas" ("status");
-- Create "do_database_users" table
CREATE TABLE "bronze"."do_database_users" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "role" character varying NULL,
  "mysql_settings_json" jsonb NULL,
  "settings_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabaseuser_cluster_id" to table: "do_database_users"
CREATE INDEX "bronzedodatabaseuser_cluster_id" ON "bronze"."do_database_users" ("cluster_id");
-- Create index "bronzedodatabaseuser_collected_at" to table: "do_database_users"
CREATE INDEX "bronzedodatabaseuser_collected_at" ON "bronze"."do_database_users" ("collected_at");
-- Create index "bronzedodatabaseuser_role" to table: "do_database_users"
CREATE INDEX "bronzedodatabaseuser_role" ON "bronze"."do_database_users" ("role");
-- Create "do_databases" table
CREATE TABLE "bronze"."do_databases" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NULL,
  "engine_slug" character varying NULL,
  "version_slug" character varying NULL,
  "num_nodes" bigint NOT NULL DEFAULT 0,
  "size_slug" character varying NULL,
  "region_slug" character varying NULL,
  "status" character varying NULL,
  "project_id" character varying NULL,
  "storage_size_mib" bigint NOT NULL DEFAULT 0,
  "private_network_uuid" character varying NULL,
  "tags_json" jsonb NULL,
  "maintenance_window_json" jsonb NULL,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodatabase_collected_at" to table: "do_databases"
CREATE INDEX "bronzedodatabase_collected_at" ON "bronze"."do_databases" ("collected_at");
-- Create index "bronzedodatabase_engine_slug" to table: "do_databases"
CREATE INDEX "bronzedodatabase_engine_slug" ON "bronze"."do_databases" ("engine_slug");
-- Create index "bronzedodatabase_project_id" to table: "do_databases"
CREATE INDEX "bronzedodatabase_project_id" ON "bronze"."do_databases" ("project_id");
-- Create index "bronzedodatabase_region_slug" to table: "do_databases"
CREATE INDEX "bronzedodatabase_region_slug" ON "bronze"."do_databases" ("region_slug");
-- Create index "bronzedodatabase_status" to table: "do_databases"
CREATE INDEX "bronzedodatabase_status" ON "bronze"."do_databases" ("status");
-- Create "do_domain_records" table
CREATE TABLE "bronze"."do_domain_records" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "domain_name" character varying NOT NULL,
  "record_id" bigint NOT NULL,
  "type" character varying NULL,
  "name" character varying NULL,
  "data" character varying NULL,
  "priority" bigint NOT NULL DEFAULT 0,
  "port" bigint NOT NULL DEFAULT 0,
  "ttl" bigint NOT NULL DEFAULT 0,
  "weight" bigint NOT NULL DEFAULT 0,
  "flags" bigint NOT NULL DEFAULT 0,
  "tag" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodomainrecord_collected_at" to table: "do_domain_records"
CREATE INDEX "bronzedodomainrecord_collected_at" ON "bronze"."do_domain_records" ("collected_at");
-- Create index "bronzedodomainrecord_domain_name" to table: "do_domain_records"
CREATE INDEX "bronzedodomainrecord_domain_name" ON "bronze"."do_domain_records" ("domain_name");
-- Create index "bronzedodomainrecord_type" to table: "do_domain_records"
CREATE INDEX "bronzedodomainrecord_type" ON "bronze"."do_domain_records" ("type");
-- Create "do_domains" table
CREATE TABLE "bronze"."do_domains" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "ttl" bigint NOT NULL DEFAULT 0,
  "zone_file" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodomain_collected_at" to table: "do_domains"
CREATE INDEX "bronzedodomain_collected_at" ON "bronze"."do_domains" ("collected_at");
-- Create "do_droplets" table
CREATE TABLE "bronze"."do_droplets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "memory" bigint NOT NULL DEFAULT 0,
  "vcpus" bigint NOT NULL DEFAULT 0,
  "disk" bigint NOT NULL DEFAULT 0,
  "region" character varying NULL,
  "size_slug" character varying NULL,
  "status" character varying NULL,
  "locked" boolean NOT NULL DEFAULT false,
  "vpc_uuid" character varying NULL,
  "api_created_at" character varying NULL,
  "image_json" jsonb NULL,
  "size_json" jsonb NULL,
  "networks_json" jsonb NULL,
  "kernel_json" jsonb NULL,
  "tags_json" jsonb NULL,
  "features_json" jsonb NULL,
  "volume_ids_json" jsonb NULL,
  "backup_ids_json" jsonb NULL,
  "snapshot_ids_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedodroplet_collected_at" to table: "do_droplets"
CREATE INDEX "bronzedodroplet_collected_at" ON "bronze"."do_droplets" ("collected_at");
-- Create index "bronzedodroplet_region" to table: "do_droplets"
CREATE INDEX "bronzedodroplet_region" ON "bronze"."do_droplets" ("region");
-- Create index "bronzedodroplet_status" to table: "do_droplets"
CREATE INDEX "bronzedodroplet_status" ON "bronze"."do_droplets" ("status");
-- Create index "bronzedodroplet_vpc_uuid" to table: "do_droplets"
CREATE INDEX "bronzedodroplet_vpc_uuid" ON "bronze"."do_droplets" ("vpc_uuid");
-- Create "do_firewalls" table
CREATE TABLE "bronze"."do_firewalls" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "status" character varying NULL,
  "inbound_rules_json" jsonb NULL,
  "outbound_rules_json" jsonb NULL,
  "droplet_ids_json" jsonb NULL,
  "tags_json" jsonb NULL,
  "api_created_at" character varying NULL,
  "pending_changes_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedofirewall_collected_at" to table: "do_firewalls"
CREATE INDEX "bronzedofirewall_collected_at" ON "bronze"."do_firewalls" ("collected_at");
-- Create index "bronzedofirewall_status" to table: "do_firewalls"
CREATE INDEX "bronzedofirewall_status" ON "bronze"."do_firewalls" ("status");
-- Create "do_keys" table
CREATE TABLE "bronze"."do_keys" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "fingerprint" character varying NULL,
  "public_key" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedokey_collected_at" to table: "do_keys"
CREATE INDEX "bronzedokey_collected_at" ON "bronze"."do_keys" ("collected_at");
-- Create index "bronzedokey_fingerprint" to table: "do_keys"
CREATE INDEX "bronzedokey_fingerprint" ON "bronze"."do_keys" ("fingerprint");
-- Create "do_kubernetes_clusters" table
CREATE TABLE "bronze"."do_kubernetes_clusters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NULL,
  "region_slug" character varying NULL,
  "version_slug" character varying NULL,
  "cluster_subnet" character varying NULL,
  "service_subnet" character varying NULL,
  "ipv4" character varying NULL,
  "endpoint" character varying NULL,
  "vpc_uuid" character varying NULL,
  "ha" boolean NOT NULL DEFAULT false,
  "auto_upgrade" boolean NOT NULL DEFAULT false,
  "surge_upgrade" boolean NOT NULL DEFAULT false,
  "registry_enabled" boolean NOT NULL DEFAULT false,
  "status_state" character varying NULL,
  "status_message" character varying NULL,
  "tags_json" jsonb NULL,
  "maintenance_policy_json" jsonb NULL,
  "control_plane_firewall_json" jsonb NULL,
  "autoscaler_config_json" jsonb NULL,
  "api_created_at" timestamptz NULL,
  "api_updated_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedokubernetescluster_collected_at" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_collected_at" ON "bronze"."do_kubernetes_clusters" ("collected_at");
-- Create index "bronzedokubernetescluster_ha" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_ha" ON "bronze"."do_kubernetes_clusters" ("ha");
-- Create index "bronzedokubernetescluster_region_slug" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_region_slug" ON "bronze"."do_kubernetes_clusters" ("region_slug");
-- Create index "bronzedokubernetescluster_status_state" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_status_state" ON "bronze"."do_kubernetes_clusters" ("status_state");
-- Create index "bronzedokubernetescluster_version_slug" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_version_slug" ON "bronze"."do_kubernetes_clusters" ("version_slug");
-- Create index "bronzedokubernetescluster_vpc_uuid" to table: "do_kubernetes_clusters"
CREATE INDEX "bronzedokubernetescluster_vpc_uuid" ON "bronze"."do_kubernetes_clusters" ("vpc_uuid");
-- Create "do_kubernetes_node_pools" table
CREATE TABLE "bronze"."do_kubernetes_node_pools" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_id" character varying NOT NULL,
  "node_pool_id" character varying NOT NULL,
  "name" character varying NULL,
  "size" character varying NULL,
  "count" bigint NOT NULL DEFAULT 0,
  "auto_scale" boolean NOT NULL DEFAULT false,
  "min_nodes" bigint NOT NULL DEFAULT 0,
  "max_nodes" bigint NOT NULL DEFAULT 0,
  "tags_json" jsonb NULL,
  "labels_json" jsonb NULL,
  "taints_json" jsonb NULL,
  "nodes_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedokubernetesnodepool_cluster_id" to table: "do_kubernetes_node_pools"
CREATE INDEX "bronzedokubernetesnodepool_cluster_id" ON "bronze"."do_kubernetes_node_pools" ("cluster_id");
-- Create index "bronzedokubernetesnodepool_collected_at" to table: "do_kubernetes_node_pools"
CREATE INDEX "bronzedokubernetesnodepool_collected_at" ON "bronze"."do_kubernetes_node_pools" ("collected_at");
-- Create index "bronzedokubernetesnodepool_size" to table: "do_kubernetes_node_pools"
CREATE INDEX "bronzedokubernetesnodepool_size" ON "bronze"."do_kubernetes_node_pools" ("size");
-- Create "do_load_balancers" table
CREATE TABLE "bronze"."do_load_balancers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NULL,
  "ip" character varying NULL,
  "ipv6" character varying NULL,
  "size_slug" character varying NULL,
  "size_unit" bigint NOT NULL DEFAULT 0,
  "lb_type" character varying NULL,
  "algorithm" character varying NULL,
  "status" character varying NULL,
  "region" character varying NULL,
  "tag" character varying NULL,
  "redirect_http_to_https" boolean NOT NULL DEFAULT false,
  "enable_proxy_protocol" boolean NOT NULL DEFAULT false,
  "enable_backend_keepalive" boolean NOT NULL DEFAULT false,
  "vpc_uuid" character varying NULL,
  "project_id" character varying NULL,
  "http_idle_timeout_seconds" bigint NULL,
  "disable_lets_encrypt_dns_records" boolean NULL,
  "network" character varying NULL,
  "network_stack" character varying NULL,
  "tls_cipher_policy" character varying NULL,
  "api_created_at" character varying NULL,
  "forwarding_rules_json" jsonb NULL,
  "health_check_json" jsonb NULL,
  "sticky_sessions_json" jsonb NULL,
  "firewall_json" jsonb NULL,
  "domains_json" jsonb NULL,
  "glb_settings_json" jsonb NULL,
  "droplet_ids_json" jsonb NULL,
  "tags_json" jsonb NULL,
  "target_load_balancer_ids_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedoloadbalancer_collected_at" to table: "do_load_balancers"
CREATE INDEX "bronzedoloadbalancer_collected_at" ON "bronze"."do_load_balancers" ("collected_at");
-- Create index "bronzedoloadbalancer_project_id" to table: "do_load_balancers"
CREATE INDEX "bronzedoloadbalancer_project_id" ON "bronze"."do_load_balancers" ("project_id");
-- Create index "bronzedoloadbalancer_region" to table: "do_load_balancers"
CREATE INDEX "bronzedoloadbalancer_region" ON "bronze"."do_load_balancers" ("region");
-- Create index "bronzedoloadbalancer_status" to table: "do_load_balancers"
CREATE INDEX "bronzedoloadbalancer_status" ON "bronze"."do_load_balancers" ("status");
-- Create index "bronzedoloadbalancer_vpc_uuid" to table: "do_load_balancers"
CREATE INDEX "bronzedoloadbalancer_vpc_uuid" ON "bronze"."do_load_balancers" ("vpc_uuid");
-- Create "do_project_resources" table
CREATE TABLE "bronze"."do_project_resources" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "project_id" character varying NOT NULL,
  "urn" character varying NOT NULL,
  "assigned_at" character varying NULL,
  "status" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedoprojectresource_collected_at" to table: "do_project_resources"
CREATE INDEX "bronzedoprojectresource_collected_at" ON "bronze"."do_project_resources" ("collected_at");
-- Create index "bronzedoprojectresource_project_id" to table: "do_project_resources"
CREATE INDEX "bronzedoprojectresource_project_id" ON "bronze"."do_project_resources" ("project_id");
-- Create index "bronzedoprojectresource_urn" to table: "do_project_resources"
CREATE INDEX "bronzedoprojectresource_urn" ON "bronze"."do_project_resources" ("urn");
-- Create "do_projects" table
CREATE TABLE "bronze"."do_projects" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "owner_uuid" character varying NULL,
  "owner_id" bigint NOT NULL DEFAULT 0,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "purpose" character varying NULL,
  "environment" character varying NULL,
  "is_default" boolean NOT NULL DEFAULT false,
  "api_created_at" character varying NULL,
  "api_updated_at" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedoproject_collected_at" to table: "do_projects"
CREATE INDEX "bronzedoproject_collected_at" ON "bronze"."do_projects" ("collected_at");
-- Create index "bronzedoproject_environment" to table: "do_projects"
CREATE INDEX "bronzedoproject_environment" ON "bronze"."do_projects" ("environment");
-- Create index "bronzedoproject_is_default" to table: "do_projects"
CREATE INDEX "bronzedoproject_is_default" ON "bronze"."do_projects" ("is_default");
-- Create "do_volumes" table
CREATE TABLE "bronze"."do_volumes" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "region" character varying NULL,
  "size_gigabytes" bigint NOT NULL DEFAULT 0,
  "description" character varying NULL,
  "droplet_ids_json" jsonb NULL,
  "filesystem_type" character varying NULL,
  "filesystem_label" character varying NULL,
  "tags_json" jsonb NULL,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedovolume_collected_at" to table: "do_volumes"
CREATE INDEX "bronzedovolume_collected_at" ON "bronze"."do_volumes" ("collected_at");
-- Create index "bronzedovolume_region" to table: "do_volumes"
CREATE INDEX "bronzedovolume_region" ON "bronze"."do_volumes" ("region");
-- Create "do_vpcs" table
CREATE TABLE "bronze"."do_vpcs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "region" character varying NULL,
  "ip_range" character varying NULL,
  "urn" character varying NULL,
  "is_default" boolean NOT NULL DEFAULT false,
  "api_created_at" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzedovpc_collected_at" to table: "do_vpcs"
CREATE INDEX "bronzedovpc_collected_at" ON "bronze"."do_vpcs" ("collected_at");
-- Create index "bronzedovpc_is_default" to table: "do_vpcs"
CREATE INDEX "bronzedovpc_is_default" ON "bronze"."do_vpcs" ("is_default");
-- Create index "bronzedovpc_region" to table: "do_vpcs"
CREATE INDEX "bronzedovpc_region" ON "bronze"."do_vpcs" ("region");
