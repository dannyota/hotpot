-- Add new schema named "silver"
CREATE SCHEMA IF NOT EXISTS "silver";
-- Create "inventory_api_endpoints" table
CREATE TABLE "silver"."inventory_api_endpoints" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "name" character varying NULL,
  "service" character varying NULL,
  "uri_pattern" character varying NOT NULL,
  "methods" jsonb NULL,
  "is_active" boolean NOT NULL DEFAULT true,
  "access_level" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventoryapiendpoint_access_level" to table: "inventory_api_endpoints"
CREATE INDEX "inventoryapiendpoint_access_level" ON "silver"."inventory_api_endpoints" ("access_level");
-- Create index "inventoryapiendpoint_service" to table: "inventory_api_endpoints"
CREATE INDEX "inventoryapiendpoint_service" ON "silver"."inventory_api_endpoints" ("service");
-- Create index "inventoryapiendpoint_uri_pattern" to table: "inventory_api_endpoints"
CREATE INDEX "inventoryapiendpoint_uri_pattern" ON "silver"."inventory_api_endpoints" ("uri_pattern");
-- Create "inventory_k8s_node_normalized" table
CREATE TABLE "silver"."inventory_k8s_node_normalized" (
  "resource_id" character varying NOT NULL,
  "provider" character varying NOT NULL,
  "is_base" boolean NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "node_name" character varying NULL,
  "cluster_name" character varying NULL,
  "node_pool" character varying NULL,
  "status" character varying NOT NULL,
  "provisioning" character varying NULL,
  "cloud_project" character varying NULL,
  "cloud_zone" character varying NULL,
  "cloud_machine_type" character varying NULL,
  "internal_ip" character varying NULL,
  "external_ip" character varying NULL,
  "merge_keys_json" jsonb NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventoryk8snodenormalized_provider" to table: "inventory_k8s_node_normalized"
CREATE INDEX "inventoryk8snodenormalized_provider" ON "silver"."inventory_k8s_node_normalized" ("provider");
-- Create index "inventoryk8snodenormalized_provider_bronze_resource_id" to table: "inventory_k8s_node_normalized"
CREATE UNIQUE INDEX "inventoryk8snodenormalized_provider_bronze_resource_id" ON "silver"."inventory_k8s_node_normalized" ("provider", "bronze_resource_id");
-- Create "inventory_machine_normalized" table
CREATE TABLE "silver"."inventory_machine_normalized" (
  "resource_id" character varying NOT NULL,
  "provider" character varying NOT NULL,
  "is_base" boolean NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "hostname" character varying NULL,
  "os_type" character varying NULL,
  "os_name" character varying NULL,
  "status" character varying NOT NULL,
  "internal_ip" character varying NULL,
  "external_ip" character varying NULL,
  "environment" character varying NULL,
  "cloud_project" character varying NULL,
  "cloud_zone" character varying NULL,
  "cloud_machine_type" character varying NULL,
  "created" timestamptz NULL,
  "merge_keys_json" jsonb NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventorymachinenormalized_provider" to table: "inventory_machine_normalized"
CREATE INDEX "inventorymachinenormalized_provider" ON "silver"."inventory_machine_normalized" ("provider");
-- Create index "inventorymachinenormalized_provider_bronze_resource_id" to table: "inventory_machine_normalized"
CREATE UNIQUE INDEX "inventorymachinenormalized_provider_bronze_resource_id" ON "silver"."inventory_machine_normalized" ("provider", "bronze_resource_id");
-- Create "inventory_software_normalized" table
CREATE TABLE "silver"."inventory_software_normalized" (
  "resource_id" character varying NOT NULL,
  "provider" character varying NOT NULL,
  "is_base" boolean NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "machine_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "version" character varying NULL,
  "publisher" character varying NULL,
  "installed_on" timestamptz NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventorysoftwarenormalized_machine_id_name" to table: "inventory_software_normalized"
CREATE INDEX "inventorysoftwarenormalized_machine_id_name" ON "silver"."inventory_software_normalized" ("machine_id", "name");
-- Create index "inventorysoftwarenormalized_provider" to table: "inventory_software_normalized"
CREATE INDEX "inventorysoftwarenormalized_provider" ON "silver"."inventory_software_normalized" ("provider");
-- Create index "inventorysoftwarenormalized_provider_bronze_resource_id" to table: "inventory_software_normalized"
CREATE UNIQUE INDEX "inventorysoftwarenormalized_provider_bronze_resource_id" ON "silver"."inventory_software_normalized" ("provider", "bronze_resource_id");
-- Create "inventory_api_endpoint_links" table
CREATE TABLE "silver"."inventory_api_endpoint_links" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "provider" character varying NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "inventory_api_endpoint_bronze_links" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_api_endpoint_links_inventory_api_endpoints_bronze_lin" FOREIGN KEY ("inventory_api_endpoint_bronze_links") REFERENCES "silver"."inventory_api_endpoints" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_k8s_nodes" table
CREATE TABLE "silver"."inventory_k8s_nodes" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "node_name" character varying NOT NULL,
  "cluster_name" character varying NOT NULL,
  "node_pool" character varying NOT NULL,
  "status" character varying NOT NULL,
  "provisioning" character varying NULL,
  "cloud_project" character varying NULL,
  "cloud_zone" character varying NULL,
  "cloud_machine_type" character varying NULL,
  "internal_ip" character varying NULL,
  "external_ip" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventoryk8snode_cluster_name" to table: "inventory_k8s_nodes"
CREATE INDEX "inventoryk8snode_cluster_name" ON "silver"."inventory_k8s_nodes" ("cluster_name");
-- Create index "inventoryk8snode_collected_at" to table: "inventory_k8s_nodes"
CREATE INDEX "inventoryk8snode_collected_at" ON "silver"."inventory_k8s_nodes" ("collected_at");
-- Create "inventory_k8s_node_links" table
CREATE TABLE "silver"."inventory_k8s_node_links" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "provider" character varying NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "inventory_k8s_node_bronze_links" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_k8s_node_links_inventory_k8s_nodes_bronze_links" FOREIGN KEY ("inventory_k8s_node_bronze_links") REFERENCES "silver"."inventory_k8s_nodes" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_machines" table
CREATE TABLE "silver"."inventory_machines" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "hostname" character varying NOT NULL,
  "os_type" character varying NOT NULL,
  "os_name" character varying NULL,
  "status" character varying NOT NULL,
  "internal_ip" character varying NULL,
  "external_ip" character varying NULL,
  "environment" character varying NULL,
  "cloud_project" character varying NULL,
  "cloud_zone" character varying NULL,
  "cloud_machine_type" character varying NULL,
  "created" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventorymachine_collected_at" to table: "inventory_machines"
CREATE INDEX "inventorymachine_collected_at" ON "silver"."inventory_machines" ("collected_at");
-- Create index "inventorymachine_environment" to table: "inventory_machines"
CREATE INDEX "inventorymachine_environment" ON "silver"."inventory_machines" ("environment");
-- Create index "inventorymachine_os_type" to table: "inventory_machines"
CREATE INDEX "inventorymachine_os_type" ON "silver"."inventory_machines" ("os_type");
-- Create "inventory_machine_links" table
CREATE TABLE "silver"."inventory_machine_links" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "provider" character varying NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "inventory_machine_bronze_links" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_machine_links_inventory_machines_bronze_links" FOREIGN KEY ("inventory_machine_bronze_links") REFERENCES "silver"."inventory_machines" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "inventory_software" table
CREATE TABLE "silver"."inventory_software" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "machine_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "version" character varying NULL,
  "publisher" character varying NULL,
  "installed_on" timestamptz NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "inventorysoftware_machine_id" to table: "inventory_software"
CREATE INDEX "inventorysoftware_machine_id" ON "silver"."inventory_software" ("machine_id");
-- Create index "inventorysoftware_machine_id_name" to table: "inventory_software"
CREATE UNIQUE INDEX "inventorysoftware_machine_id_name" ON "silver"."inventory_software" ("machine_id", "name");
-- Create index "inventorysoftware_name" to table: "inventory_software"
CREATE INDEX "inventorysoftware_name" ON "silver"."inventory_software" ("name");
-- Create "inventory_software_links" table
CREATE TABLE "silver"."inventory_software_links" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "provider" character varying NOT NULL,
  "bronze_table" character varying NOT NULL,
  "bronze_resource_id" character varying NOT NULL,
  "inventory_software_bronze_links" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "inventory_software_links_inventory_software_bronze_links" FOREIGN KEY ("inventory_software_bronze_links") REFERENCES "silver"."inventory_software" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
