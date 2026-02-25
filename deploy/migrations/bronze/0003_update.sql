-- Modify "greennode_network_endpoints" table
ALTER TABLE "bronze"."greennode_network_endpoints" ADD COLUMN "endpoint_auth_url" character varying NULL, ADD COLUMN "endpoint_service_id" character varying NULL, ADD COLUMN "billing_status" character varying NULL, ADD COLUMN "endpoint_type" character varying NULL, ADD COLUMN "version" character varying NULL, ADD COLUMN "description" character varying NULL, ADD COLUMN "created_at" character varying NULL, ADD COLUMN "updated_at" character varying NULL, ADD COLUMN "vpc_name" character varying NULL, ADD COLUMN "zone_uuid" character varying NULL, ADD COLUMN "enable_dns_name" boolean NOT NULL DEFAULT false, ADD COLUMN "endpoint_domains" jsonb NULL, ADD COLUMN "subnet_id" character varying NULL, ADD COLUMN "category_name" character varying NULL, ADD COLUMN "service_name" character varying NULL, ADD COLUMN "service_endpoint_type" character varying NULL, ADD COLUMN "package_name" character varying NULL;
-- Modify "greennode_network_secgroups" table
ALTER TABLE "bronze"."greennode_network_secgroups" ADD COLUMN "created_at" character varying NULL, ADD COLUMN "is_system" boolean NOT NULL DEFAULT false;
-- Create "greennode_compute_os_images" table
CREATE TABLE "bronze"."greennode_compute_os_images" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "image_type" character varying NULL,
  "image_version" character varying NULL,
  "licence" boolean NULL,
  "license_key" character varying NULL,
  "description" character varying NULL,
  "zone_id" character varying NULL,
  "flavor_zone_ids" jsonb NULL,
  "default_tag_ids" jsonb NULL,
  "package_limit_cpu" bigint NULL,
  "package_limit_memory" bigint NULL,
  "package_limit_disk_size" bigint NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodecomputeosimage_collected_at" to table: "greennode_compute_os_images"
CREATE INDEX "bronzegreennodecomputeosimage_collected_at" ON "bronze"."greennode_compute_os_images" ("collected_at");
-- Create index "bronzegreennodecomputeosimage_project_id" to table: "greennode_compute_os_images"
CREATE INDEX "bronzegreennodecomputeosimage_project_id" ON "bronze"."greennode_compute_os_images" ("project_id");
-- Create index "bronzegreennodecomputeosimage_region" to table: "greennode_compute_os_images"
CREATE INDEX "bronzegreennodecomputeosimage_region" ON "bronze"."greennode_compute_os_images" ("region");
-- Create "greennode_compute_user_images" table
CREATE TABLE "bronze"."greennode_compute_user_images" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "status" character varying NULL,
  "min_disk" bigint NULL,
  "image_size" double precision NULL,
  "meta_data" character varying NULL,
  "created_at" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodecomputeuserimage_collected_at" to table: "greennode_compute_user_images"
CREATE INDEX "bronzegreennodecomputeuserimage_collected_at" ON "bronze"."greennode_compute_user_images" ("collected_at");
-- Create index "bronzegreennodecomputeuserimage_project_id" to table: "greennode_compute_user_images"
CREATE INDEX "bronzegreennodecomputeuserimage_project_id" ON "bronze"."greennode_compute_user_images" ("project_id");
-- Create index "bronzegreennodecomputeuserimage_region" to table: "greennode_compute_user_images"
CREATE INDEX "bronzegreennodecomputeuserimage_region" ON "bronze"."greennode_compute_user_images" ("region");
-- Create index "bronzegreennodecomputeuserimage_status" to table: "greennode_compute_user_images"
CREATE INDEX "bronzegreennodecomputeuserimage_status" ON "bronze"."greennode_compute_user_images" ("status");
-- Create "greennode_network_interconnects" table
CREATE TABLE "bronze"."greennode_network_interconnects" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "enable_gw2" boolean NOT NULL DEFAULT false,
  "circuit_id" bigint NULL,
  "gw01_ip" character varying NULL,
  "gw02_ip" character varying NULL,
  "gw_vip" character varying NULL,
  "remote_gw01_ip" character varying NULL,
  "remote_gw02_ip" character varying NULL,
  "package_id" character varying NULL,
  "type_id" character varying NULL,
  "type_name" character varying NULL,
  "created_at" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworkinterconnect_collected_at" to table: "greennode_network_interconnects"
CREATE INDEX "bronzegreennodenetworkinterconnect_collected_at" ON "bronze"."greennode_network_interconnects" ("collected_at");
-- Create index "bronzegreennodenetworkinterconnect_project_id" to table: "greennode_network_interconnects"
CREATE INDEX "bronzegreennodenetworkinterconnect_project_id" ON "bronze"."greennode_network_interconnects" ("project_id");
-- Create index "bronzegreennodenetworkinterconnect_region" to table: "greennode_network_interconnects"
CREATE INDEX "bronzegreennodenetworkinterconnect_region" ON "bronze"."greennode_network_interconnects" ("region");
-- Create index "bronzegreennodenetworkinterconnect_status" to table: "greennode_network_interconnects"
CREATE INDEX "bronzegreennodenetworkinterconnect_status" ON "bronze"."greennode_network_interconnects" ("status");
-- Create "greennode_network_peerings" table
CREATE TABLE "bronze"."greennode_network_peerings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "status" character varying NULL,
  "from_vpc_id" character varying NULL,
  "from_cidr" character varying NULL,
  "end_vpc_id" character varying NULL,
  "end_cidr" character varying NULL,
  "created_at" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworkpeering_collected_at" to table: "greennode_network_peerings"
CREATE INDEX "bronzegreennodenetworkpeering_collected_at" ON "bronze"."greennode_network_peerings" ("collected_at");
-- Create index "bronzegreennodenetworkpeering_project_id" to table: "greennode_network_peerings"
CREATE INDEX "bronzegreennodenetworkpeering_project_id" ON "bronze"."greennode_network_peerings" ("project_id");
-- Create index "bronzegreennodenetworkpeering_region" to table: "greennode_network_peerings"
CREATE INDEX "bronzegreennodenetworkpeering_region" ON "bronze"."greennode_network_peerings" ("region");
-- Create index "bronzegreennodenetworkpeering_status" to table: "greennode_network_peerings"
CREATE INDEX "bronzegreennodenetworkpeering_status" ON "bronze"."greennode_network_peerings" ("status");
-- Create "greennode_network_route_tables" table
CREATE TABLE "bronze"."greennode_network_route_tables" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "status" character varying NULL,
  "network_id" character varying NULL,
  "created_at" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworkroutetable_collected_at" to table: "greennode_network_route_tables"
CREATE INDEX "bronzegreennodenetworkroutetable_collected_at" ON "bronze"."greennode_network_route_tables" ("collected_at");
-- Create index "bronzegreennodenetworkroutetable_project_id" to table: "greennode_network_route_tables"
CREATE INDEX "bronzegreennodenetworkroutetable_project_id" ON "bronze"."greennode_network_route_tables" ("project_id");
-- Create index "bronzegreennodenetworkroutetable_region" to table: "greennode_network_route_tables"
CREATE INDEX "bronzegreennodenetworkroutetable_region" ON "bronze"."greennode_network_route_tables" ("region");
-- Create index "bronzegreennodenetworkroutetable_status" to table: "greennode_network_route_tables"
CREATE INDEX "bronzegreennodenetworkroutetable_status" ON "bronze"."greennode_network_route_tables" ("status");
-- Create "greennode_network_subnets" table
CREATE TABLE "bronze"."greennode_network_subnets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "network_id" character varying NOT NULL,
  "cidr" character varying NULL,
  "status" character varying NULL,
  "route_table_id" character varying NULL,
  "interface_acl_policy_id" character varying NULL,
  "interface_acl_policy_name" character varying NULL,
  "zone_id" character varying NULL,
  "secondary_subnets" jsonb NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworksubnet_collected_at" to table: "greennode_network_subnets"
CREATE INDEX "bronzegreennodenetworksubnet_collected_at" ON "bronze"."greennode_network_subnets" ("collected_at");
-- Create index "bronzegreennodenetworksubnet_network_id" to table: "greennode_network_subnets"
CREATE INDEX "bronzegreennodenetworksubnet_network_id" ON "bronze"."greennode_network_subnets" ("network_id");
-- Create index "bronzegreennodenetworksubnet_project_id" to table: "greennode_network_subnets"
CREATE INDEX "bronzegreennodenetworksubnet_project_id" ON "bronze"."greennode_network_subnets" ("project_id");
-- Create index "bronzegreennodenetworksubnet_region" to table: "greennode_network_subnets"
CREATE INDEX "bronzegreennodenetworksubnet_region" ON "bronze"."greennode_network_subnets" ("region");
-- Create index "bronzegreennodenetworksubnet_status" to table: "greennode_network_subnets"
CREATE INDEX "bronzegreennodenetworksubnet_status" ON "bronze"."greennode_network_subnets" ("status");
-- Create "greennode_network_vpcs" table
CREATE TABLE "bronze"."greennode_network_vpcs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "cidr" character varying NULL,
  "status" character varying NULL,
  "route_table_id" character varying NULL,
  "route_table_name" character varying NULL,
  "dhcp_option_id" character varying NULL,
  "dhcp_option_name" character varying NULL,
  "dns_status" character varying NULL,
  "dns_id" character varying NULL,
  "zone_uuid" character varying NULL,
  "zone_name" character varying NULL,
  "created_at" character varying NULL,
  "elastic_ips" jsonb NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworkvpc_collected_at" to table: "greennode_network_vpcs"
CREATE INDEX "bronzegreennodenetworkvpc_collected_at" ON "bronze"."greennode_network_vpcs" ("collected_at");
-- Create index "bronzegreennodenetworkvpc_project_id" to table: "greennode_network_vpcs"
CREATE INDEX "bronzegreennodenetworkvpc_project_id" ON "bronze"."greennode_network_vpcs" ("project_id");
-- Create index "bronzegreennodenetworkvpc_region" to table: "greennode_network_vpcs"
CREATE INDEX "bronzegreennodenetworkvpc_region" ON "bronze"."greennode_network_vpcs" ("region");
-- Create index "bronzegreennodenetworkvpc_status" to table: "greennode_network_vpcs"
CREATE INDEX "bronzegreennodenetworkvpc_status" ON "bronze"."greennode_network_vpcs" ("status");
-- Create "greennode_network_route_table_routes" table
CREATE TABLE "bronze"."greennode_network_route_table_routes" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "route_id" character varying NOT NULL,
  "routing_type" character varying NULL,
  "destination_cidr_block" character varying NULL,
  "target" character varying NULL,
  "status" character varying NULL,
  "bronze_green_node_network_route_table_routes" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_network_route_table_routes_greennode_network_route_ta" FOREIGN KEY ("bronze_green_node_network_route_table_routes") REFERENCES "bronze"."greennode_network_route_tables" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Drop "s1_threats" table
DROP TABLE "bronze"."s1_threats";
