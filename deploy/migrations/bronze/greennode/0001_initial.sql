-- Create "greennode_volume_volume_types" table
CREATE TABLE "bronze"."greennode_volume_volume_types" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "iops" bigint NULL,
  "max_size" bigint NULL,
  "min_size" bigint NULL,
  "through_put" bigint NULL,
  "zone_id" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodevolumevolumetype_collected_at" to table: "greennode_volume_volume_types"
CREATE INDEX "bronzegreennodevolumevolumetype_collected_at" ON "bronze"."greennode_volume_volume_types" ("collected_at");
-- Create index "bronzegreennodevolumevolumetype_project_id" to table: "greennode_volume_volume_types"
CREATE INDEX "bronzegreennodevolumevolumetype_project_id" ON "bronze"."greennode_volume_volume_types" ("project_id");
-- Create index "bronzegreennodevolumevolumetype_region" to table: "greennode_volume_volume_types"
CREATE INDEX "bronzegreennodevolumevolumetype_region" ON "bronze"."greennode_volume_volume_types" ("region");
-- Create "greennode_network_endpoints" table
CREATE TABLE "bronze"."greennode_network_endpoints" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "ipv4_address" character varying NULL,
  "endpoint_url" character varying NULL,
  "endpoint_auth_url" character varying NULL,
  "endpoint_service_id" character varying NULL,
  "status" character varying NULL,
  "billing_status" character varying NULL,
  "endpoint_type" character varying NULL,
  "version" character varying NULL,
  "description" character varying NULL,
  "created_at" character varying NULL,
  "updated_at" character varying NULL,
  "vpc_id" character varying NULL,
  "vpc_name" character varying NULL,
  "zone_uuid" character varying NULL,
  "enable_dns_name" boolean NOT NULL DEFAULT false,
  "endpoint_domains" jsonb NULL,
  "subnet_id" character varying NULL,
  "category_name" character varying NULL,
  "service_name" character varying NULL,
  "service_endpoint_type" character varying NULL,
  "package_name" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworkendpoint_collected_at" to table: "greennode_network_endpoints"
CREATE INDEX "bronzegreennodenetworkendpoint_collected_at" ON "bronze"."greennode_network_endpoints" ("collected_at");
-- Create index "bronzegreennodenetworkendpoint_project_id" to table: "greennode_network_endpoints"
CREATE INDEX "bronzegreennodenetworkendpoint_project_id" ON "bronze"."greennode_network_endpoints" ("project_id");
-- Create index "bronzegreennodenetworkendpoint_region" to table: "greennode_network_endpoints"
CREATE INDEX "bronzegreennodenetworkendpoint_region" ON "bronze"."greennode_network_endpoints" ("region");
-- Create index "bronzegreennodenetworkendpoint_status" to table: "greennode_network_endpoints"
CREATE INDEX "bronzegreennodenetworkendpoint_status" ON "bronze"."greennode_network_endpoints" ("status");
-- Create "greennode_compute_server_groups" table
CREATE TABLE "bronze"."greennode_compute_server_groups" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "policy_id" character varying NULL,
  "policy_name" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodecomputeservergroup_collected_at" to table: "greennode_compute_server_groups"
CREATE INDEX "bronzegreennodecomputeservergroup_collected_at" ON "bronze"."greennode_compute_server_groups" ("collected_at");
-- Create index "bronzegreennodecomputeservergroup_project_id" to table: "greennode_compute_server_groups"
CREATE INDEX "bronzegreennodecomputeservergroup_project_id" ON "bronze"."greennode_compute_server_groups" ("project_id");
-- Create index "bronzegreennodecomputeservergroup_region" to table: "greennode_compute_server_groups"
CREATE INDEX "bronzegreennodecomputeservergroup_region" ON "bronze"."greennode_compute_server_groups" ("region");
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
-- Create "greennode_portal_quotas" table
CREATE TABLE "bronze"."greennode_portal_quotas" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "type" character varying NULL,
  "limit_value" bigint NOT NULL,
  "used_value" bigint NOT NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeportalquota_collected_at" to table: "greennode_portal_quotas"
CREATE INDEX "bronzegreennodeportalquota_collected_at" ON "bronze"."greennode_portal_quotas" ("collected_at");
-- Create index "bronzegreennodeportalquota_project_id" to table: "greennode_portal_quotas"
CREATE INDEX "bronzegreennodeportalquota_project_id" ON "bronze"."greennode_portal_quotas" ("project_id");
-- Create index "bronzegreennodeportalquota_region" to table: "greennode_portal_quotas"
CREATE INDEX "bronzegreennodeportalquota_region" ON "bronze"."greennode_portal_quotas" ("region");
-- Create "greennode_compute_ssh_keys" table
CREATE TABLE "bronze"."greennode_compute_ssh_keys" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "created_at_api" character varying NULL,
  "pub_key" character varying NULL,
  "status" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodecomputesshkey_collected_at" to table: "greennode_compute_ssh_keys"
CREATE INDEX "bronzegreennodecomputesshkey_collected_at" ON "bronze"."greennode_compute_ssh_keys" ("collected_at");
-- Create index "bronzegreennodecomputesshkey_project_id" to table: "greennode_compute_ssh_keys"
CREATE INDEX "bronzegreennodecomputesshkey_project_id" ON "bronze"."greennode_compute_ssh_keys" ("project_id");
-- Create index "bronzegreennodecomputesshkey_region" to table: "greennode_compute_ssh_keys"
CREATE INDEX "bronzegreennodecomputesshkey_region" ON "bronze"."greennode_compute_ssh_keys" ("region");
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
-- Create "greennode_glb_global_regions" table
CREATE TABLE "bronze"."greennode_glb_global_regions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "vserver_endpoint" character varying NULL,
  "vlb_endpoint" character varying NULL,
  "ui_server_endpoint" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeglbglobalregion_collected_at" to table: "greennode_glb_global_regions"
CREATE INDEX "bronzegreennodeglbglobalregion_collected_at" ON "bronze"."greennode_glb_global_regions" ("collected_at");
-- Create index "bronzegreennodeglbglobalregion_project_id" to table: "greennode_glb_global_regions"
CREATE INDEX "bronzegreennodeglbglobalregion_project_id" ON "bronze"."greennode_glb_global_regions" ("project_id");
-- Create "greennode_portal_zones" table
CREATE TABLE "bronze"."greennode_portal_zones" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "openstack_zone" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeportalzone_collected_at" to table: "greennode_portal_zones"
CREATE INDEX "bronzegreennodeportalzone_collected_at" ON "bronze"."greennode_portal_zones" ("collected_at");
-- Create index "bronzegreennodeportalzone_project_id" to table: "greennode_portal_zones"
CREATE INDEX "bronzegreennodeportalzone_project_id" ON "bronze"."greennode_portal_zones" ("project_id");
-- Create "greennode_volume_volume_type_zones" table
CREATE TABLE "bronze"."greennode_volume_volume_type_zones" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "pool_name_json" jsonb NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodevolumevolumetypezone_collected_at" to table: "greennode_volume_volume_type_zones"
CREATE INDEX "bronzegreennodevolumevolumetypezone_collected_at" ON "bronze"."greennode_volume_volume_type_zones" ("collected_at");
-- Create index "bronzegreennodevolumevolumetypezone_project_id" to table: "greennode_volume_volume_type_zones"
CREATE INDEX "bronzegreennodevolumevolumetypezone_project_id" ON "bronze"."greennode_volume_volume_type_zones" ("project_id");
-- Create index "bronzegreennodevolumevolumetypezone_region" to table: "greennode_volume_volume_type_zones"
CREATE INDEX "bronzegreennodevolumevolumetypezone_region" ON "bronze"."greennode_volume_volume_type_zones" ("region");
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
-- Create "greennode_loadbalancer_certificates" table
CREATE TABLE "bronze"."greennode_loadbalancer_certificates" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "certificate_type" character varying NULL,
  "expired_at" character varying NULL,
  "imported_at" character varying NULL,
  "not_after" bigint NULL,
  "key_algorithm" character varying NULL,
  "serial" character varying NULL,
  "subject" character varying NULL,
  "domain_name" character varying NULL,
  "in_use" boolean NOT NULL DEFAULT false,
  "issuer" character varying NULL,
  "signature_algorithm" character varying NULL,
  "not_before" bigint NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeloadbalancercertificate_collected_at" to table: "greennode_loadbalancer_certificates"
CREATE INDEX "bronzegreennodeloadbalancercertificate_collected_at" ON "bronze"."greennode_loadbalancer_certificates" ("collected_at");
-- Create index "bronzegreennodeloadbalancercertificate_project_id" to table: "greennode_loadbalancer_certificates"
CREATE INDEX "bronzegreennodeloadbalancercertificate_project_id" ON "bronze"."greennode_loadbalancer_certificates" ("project_id");
-- Create index "bronzegreennodeloadbalancercertificate_region" to table: "greennode_loadbalancer_certificates"
CREATE INDEX "bronzegreennodeloadbalancercertificate_region" ON "bronze"."greennode_loadbalancer_certificates" ("region");
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
-- Create "greennode_portal_regions" table
CREATE TABLE "bronze"."greennode_portal_regions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeportalregion_collected_at" to table: "greennode_portal_regions"
CREATE INDEX "bronzegreennodeportalregion_collected_at" ON "bronze"."greennode_portal_regions" ("collected_at");
-- Create index "bronzegreennodeportalregion_project_id" to table: "greennode_portal_regions"
CREATE INDEX "bronzegreennodeportalregion_project_id" ON "bronze"."greennode_portal_regions" ("project_id");
-- Create "greennode_glb_global_packages" table
CREATE TABLE "bronze"."greennode_glb_global_packages" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "description_en" character varying NULL,
  "detail_json" jsonb NULL,
  "enabled" boolean NOT NULL DEFAULT false,
  "base_sku" character varying NULL,
  "base_connection_rate" bigint NULL,
  "base_domestic_traffic_total" bigint NULL,
  "base_non_domestic_traffic_total" bigint NULL,
  "connection_sku" character varying NULL,
  "domestic_traffic_sku" character varying NULL,
  "non_domestic_traffic_sku" character varying NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "vlb_packages_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeglbglobalpackage_collected_at" to table: "greennode_glb_global_packages"
CREATE INDEX "bronzegreennodeglbglobalpackage_collected_at" ON "bronze"."greennode_glb_global_packages" ("collected_at");
-- Create index "bronzegreennodeglbglobalpackage_project_id" to table: "greennode_glb_global_packages"
CREATE INDEX "bronzegreennodeglbglobalpackage_project_id" ON "bronze"."greennode_glb_global_packages" ("project_id");
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
-- Create "greennode_loadbalancer_packages" table
CREATE TABLE "bronze"."greennode_loadbalancer_packages" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "type" character varying NULL,
  "connection_number" bigint NULL,
  "data_transfer" bigint NULL,
  "mode" character varying NULL,
  "lb_type" character varying NULL,
  "display_lb_type" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeloadbalancerpackage_collected_at" to table: "greennode_loadbalancer_packages"
CREATE INDEX "bronzegreennodeloadbalancerpackage_collected_at" ON "bronze"."greennode_loadbalancer_packages" ("collected_at");
-- Create index "bronzegreennodeloadbalancerpackage_project_id" to table: "greennode_loadbalancer_packages"
CREATE INDEX "bronzegreennodeloadbalancerpackage_project_id" ON "bronze"."greennode_loadbalancer_packages" ("project_id");
-- Create index "bronzegreennodeloadbalancerpackage_region" to table: "greennode_loadbalancer_packages"
CREATE INDEX "bronzegreennodeloadbalancerpackage_region" ON "bronze"."greennode_loadbalancer_packages" ("region");
-- Create "greennode_compute_server_group_members" table
CREATE TABLE "bronze"."greennode_compute_server_group_members" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "uuid" character varying NOT NULL,
  "name" character varying NOT NULL,
  "bronze_green_node_compute_server_group_members" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_compute_server_group_members_greennode_compute_server" FOREIGN KEY ("bronze_green_node_compute_server_group_members") REFERENCES "bronze"."greennode_compute_server_groups" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_compute_servers" table
CREATE TABLE "bronze"."greennode_compute_servers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "status" character varying NULL,
  "location" character varying NULL,
  "zone_id" character varying NULL,
  "created_at_api" character varying NULL,
  "boot_volume_id" character varying NULL,
  "encryption_volume" boolean NOT NULL DEFAULT false,
  "licence" boolean NOT NULL DEFAULT false,
  "metadata" character varying NULL,
  "migrate_state" character varying NULL,
  "product" character varying NULL,
  "server_group_id" character varying NULL,
  "server_group_name" character varying NULL,
  "ssh_key_name" character varying NULL,
  "stop_before_migrate" boolean NOT NULL DEFAULT false,
  "user" character varying NULL,
  "image_id" character varying NULL,
  "image_type" character varying NULL,
  "image_version" character varying NULL,
  "flavor_id" character varying NULL,
  "flavor_name" character varying NULL,
  "flavor_cpu" bigint NULL,
  "flavor_memory" bigint NULL,
  "flavor_gpu" bigint NULL,
  "flavor_bandwidth" bigint NULL,
  "interfaces_json" jsonb NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodecomputeserver_collected_at" to table: "greennode_compute_servers"
CREATE INDEX "bronzegreennodecomputeserver_collected_at" ON "bronze"."greennode_compute_servers" ("collected_at");
-- Create index "bronzegreennodecomputeserver_project_id" to table: "greennode_compute_servers"
CREATE INDEX "bronzegreennodecomputeserver_project_id" ON "bronze"."greennode_compute_servers" ("project_id");
-- Create index "bronzegreennodecomputeserver_region" to table: "greennode_compute_servers"
CREATE INDEX "bronzegreennodecomputeserver_region" ON "bronze"."greennode_compute_servers" ("region");
-- Create index "bronzegreennodecomputeserver_status" to table: "greennode_compute_servers"
CREATE INDEX "bronzegreennodecomputeserver_status" ON "bronze"."greennode_compute_servers" ("status");
-- Create "greennode_compute_server_sec_groups" table
CREATE TABLE "bronze"."greennode_compute_server_sec_groups" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "uuid" character varying NOT NULL,
  "name" character varying NOT NULL,
  "bronze_green_node_compute_server_sec_groups" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_compute_server_sec_groups_greennode_compute_servers_s" FOREIGN KEY ("bronze_green_node_compute_server_sec_groups") REFERENCES "bronze"."greennode_compute_servers" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_dns_hosted_zones" table
CREATE TABLE "bronze"."greennode_dns_hosted_zones" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "domain_name" character varying NOT NULL,
  "status" character varying NULL,
  "description" character varying NULL,
  "type" character varying NULL,
  "count_records" bigint NULL,
  "assoc_vpc_ids_json" jsonb NULL,
  "assoc_vpc_map_region_json" jsonb NULL,
  "portal_user_id" bigint NULL,
  "created_at_api" character varying NULL,
  "deleted_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodednshostedzone_collected_at" to table: "greennode_dns_hosted_zones"
CREATE INDEX "bronzegreennodednshostedzone_collected_at" ON "bronze"."greennode_dns_hosted_zones" ("collected_at");
-- Create index "bronzegreennodednshostedzone_project_id" to table: "greennode_dns_hosted_zones"
CREATE INDEX "bronzegreennodednshostedzone_project_id" ON "bronze"."greennode_dns_hosted_zones" ("project_id");
-- Create index "bronzegreennodednshostedzone_status" to table: "greennode_dns_hosted_zones"
CREATE INDEX "bronzegreennodednshostedzone_status" ON "bronze"."greennode_dns_hosted_zones" ("status");
-- Create "greennode_dns_records" table
CREATE TABLE "bronze"."greennode_dns_records" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "record_id" character varying NOT NULL,
  "sub_domain" character varying NULL,
  "status" character varying NULL,
  "type" character varying NULL,
  "routing_policy" character varying NULL,
  "value_json" jsonb NULL,
  "ttl" bigint NULL,
  "enable_sticky_session" boolean NULL,
  "created_at_api" character varying NULL,
  "deleted_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "bronze_green_node_dns_hosted_zone_records" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_dns_records_greennode_dns_hosted_zones_records" FOREIGN KEY ("bronze_green_node_dns_hosted_zone_records") REFERENCES "bronze"."greennode_dns_hosted_zones" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_glb_global_load_balancers" table
CREATE TABLE "bronze"."greennode_glb_global_load_balancers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "package" character varying NULL,
  "type" character varying NULL,
  "user_id" bigint NULL,
  "vips_json" jsonb NULL,
  "domains_json" jsonb NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "deleted_at_api" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeglbgloballoadbalancer_collected_at" to table: "greennode_glb_global_load_balancers"
CREATE INDEX "bronzegreennodeglbgloballoadbalancer_collected_at" ON "bronze"."greennode_glb_global_load_balancers" ("collected_at");
-- Create index "bronzegreennodeglbgloballoadbalancer_project_id" to table: "greennode_glb_global_load_balancers"
CREATE INDEX "bronzegreennodeglbgloballoadbalancer_project_id" ON "bronze"."greennode_glb_global_load_balancers" ("project_id");
-- Create index "bronzegreennodeglbgloballoadbalancer_status" to table: "greennode_glb_global_load_balancers"
CREATE INDEX "bronzegreennodeglbgloballoadbalancer_status" ON "bronze"."greennode_glb_global_load_balancers" ("status");
-- Create "greennode_glb_global_listeners" table
CREATE TABLE "bronze"."greennode_glb_global_listeners" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "listener_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "protocol" character varying NULL,
  "port" bigint NULL,
  "global_pool_id" character varying NULL,
  "timeout_client" bigint NULL,
  "timeout_member" bigint NULL,
  "timeout_connection" bigint NULL,
  "allowed_cidrs" character varying NULL,
  "headers" character varying NULL,
  "status" character varying NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "deleted_at_api" character varying NULL,
  "bronze_green_node_glb_global_load_balancer_listeners" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_glb_global_listeners_greennode_glb_global_load_balanc" FOREIGN KEY ("bronze_green_node_glb_global_load_balancer_listeners") REFERENCES "bronze"."greennode_glb_global_load_balancers" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_glb_global_pools" table
CREATE TABLE "bronze"."greennode_glb_global_pools" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "pool_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "algorithm" character varying NULL,
  "sticky_session" character varying NULL,
  "tls_enabled" character varying NULL,
  "protocol" character varying NULL,
  "status" character varying NULL,
  "health_json" jsonb NULL,
  "pool_members_json" jsonb NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "deleted_at_api" character varying NULL,
  "bronze_green_node_glb_global_load_balancer_pools" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_glb_global_pools_greennode_glb_global_load_balancers_" FOREIGN KEY ("bronze_green_node_glb_global_load_balancer_pools") REFERENCES "bronze"."greennode_glb_global_load_balancers" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_loadbalancer_lbs" table
CREATE TABLE "bronze"."greennode_loadbalancer_lbs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_status" character varying NULL,
  "address" character varying NULL,
  "private_subnet_id" character varying NULL,
  "private_subnet_cidr" character varying NULL,
  "type" character varying NULL,
  "display_type" character varying NULL,
  "load_balancer_schema" character varying NULL,
  "package_id" character varying NULL,
  "description" character varying NULL,
  "location" character varying NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "progress_status" character varying NULL,
  "status" character varying NULL,
  "backend_subnet_id" character varying NULL,
  "internal" boolean NOT NULL DEFAULT false,
  "auto_scalable" boolean NOT NULL DEFAULT false,
  "zone_id" character varying NULL,
  "min_size" bigint NULL,
  "max_size" bigint NULL,
  "total_nodes" bigint NULL,
  "nodes_json" jsonb NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodeloadbalancerlb_collected_at" to table: "greennode_loadbalancer_lbs"
CREATE INDEX "bronzegreennodeloadbalancerlb_collected_at" ON "bronze"."greennode_loadbalancer_lbs" ("collected_at");
-- Create index "bronzegreennodeloadbalancerlb_project_id" to table: "greennode_loadbalancer_lbs"
CREATE INDEX "bronzegreennodeloadbalancerlb_project_id" ON "bronze"."greennode_loadbalancer_lbs" ("project_id");
-- Create index "bronzegreennodeloadbalancerlb_region" to table: "greennode_loadbalancer_lbs"
CREATE INDEX "bronzegreennodeloadbalancerlb_region" ON "bronze"."greennode_loadbalancer_lbs" ("region");
-- Create index "bronzegreennodeloadbalancerlb_status" to table: "greennode_loadbalancer_lbs"
CREATE INDEX "bronzegreennodeloadbalancerlb_status" ON "bronze"."greennode_loadbalancer_lbs" ("status");
-- Create "greennode_loadbalancer_listeners" table
CREATE TABLE "bronze"."greennode_loadbalancer_listeners" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "listener_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "protocol" character varying NULL,
  "protocol_port" bigint NULL,
  "connection_limit" bigint NULL,
  "default_pool_id" character varying NULL,
  "default_pool_name" character varying NULL,
  "timeout_client" bigint NULL,
  "timeout_member" bigint NULL,
  "timeout_connection" bigint NULL,
  "allowed_cidrs" character varying NULL,
  "certificate_authorities_json" jsonb NULL,
  "display_status" character varying NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "default_certificate_authority" character varying NULL,
  "client_certificate_authentication" character varying NULL,
  "progress_status" character varying NULL,
  "insert_headers_json" jsonb NULL,
  "policies_json" jsonb NULL,
  "bronze_green_node_load_balancer_lb_listeners" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_loadbalancer_listeners_greennode_loadbalancer_lbs_lis" FOREIGN KEY ("bronze_green_node_load_balancer_lb_listeners") REFERENCES "bronze"."greennode_loadbalancer_lbs" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_loadbalancer_pools" table
CREATE TABLE "bronze"."greennode_loadbalancer_pools" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "pool_id" character varying NOT NULL,
  "name" character varying NOT NULL,
  "protocol" character varying NULL,
  "description" character varying NULL,
  "load_balance_method" character varying NULL,
  "status" character varying NULL,
  "stickiness" boolean NOT NULL DEFAULT false,
  "tls_encryption" boolean NOT NULL DEFAULT false,
  "members_json" jsonb NULL,
  "health_monitor_json" jsonb NULL,
  "bronze_green_node_load_balancer_lb_pools" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_loadbalancer_pools_greennode_loadbalancer_lbs_pools" FOREIGN KEY ("bronze_green_node_load_balancer_lb_pools") REFERENCES "bronze"."greennode_loadbalancer_lbs" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
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
-- Create "greennode_network_secgroups" table
CREATE TABLE "bronze"."greennode_network_secgroups" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "created_at" character varying NULL,
  "is_system" boolean NOT NULL DEFAULT false,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodenetworksecgroup_collected_at" to table: "greennode_network_secgroups"
CREATE INDEX "bronzegreennodenetworksecgroup_collected_at" ON "bronze"."greennode_network_secgroups" ("collected_at");
-- Create index "bronzegreennodenetworksecgroup_project_id" to table: "greennode_network_secgroups"
CREATE INDEX "bronzegreennodenetworksecgroup_project_id" ON "bronze"."greennode_network_secgroups" ("project_id");
-- Create index "bronzegreennodenetworksecgroup_region" to table: "greennode_network_secgroups"
CREATE INDEX "bronzegreennodenetworksecgroup_region" ON "bronze"."greennode_network_secgroups" ("region");
-- Create index "bronzegreennodenetworksecgroup_status" to table: "greennode_network_secgroups"
CREATE INDEX "bronzegreennodenetworksecgroup_status" ON "bronze"."greennode_network_secgroups" ("status");
-- Create "greennode_network_secgroup_rules" table
CREATE TABLE "bronze"."greennode_network_secgroup_rules" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "rule_id" character varying NOT NULL,
  "direction" character varying NULL,
  "ether_type" character varying NULL,
  "protocol" character varying NULL,
  "description" character varying NULL,
  "remote_ip_prefix" character varying NULL,
  "port_range_max" bigint NULL,
  "port_range_min" bigint NULL,
  "bronze_green_node_network_secgroup_rules" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_network_secgroup_rules_greennode_network_secgroups_ru" FOREIGN KEY ("bronze_green_node_network_secgroup_rules") REFERENCES "bronze"."greennode_network_secgroups" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "greennode_volume_block_volumes" table
CREATE TABLE "bronze"."greennode_volume_block_volumes" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "volume_type_id" character varying NULL,
  "cluster_id" character varying NULL,
  "vm_id" character varying NULL,
  "size" character varying NULL,
  "iops_id" character varying NULL,
  "status" character varying NULL,
  "created_at_api" character varying NULL,
  "updated_at_api" character varying NULL,
  "persistent_volume" boolean NOT NULL DEFAULT false,
  "attached_machine_json" jsonb NULL,
  "under_id" character varying NULL,
  "migrate_state" character varying NULL,
  "multi_attach" boolean NOT NULL DEFAULT false,
  "zone_id" character varying NULL,
  "region" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegreennodevolumeblockvolume_collected_at" to table: "greennode_volume_block_volumes"
CREATE INDEX "bronzegreennodevolumeblockvolume_collected_at" ON "bronze"."greennode_volume_block_volumes" ("collected_at");
-- Create index "bronzegreennodevolumeblockvolume_project_id" to table: "greennode_volume_block_volumes"
CREATE INDEX "bronzegreennodevolumeblockvolume_project_id" ON "bronze"."greennode_volume_block_volumes" ("project_id");
-- Create index "bronzegreennodevolumeblockvolume_region" to table: "greennode_volume_block_volumes"
CREATE INDEX "bronzegreennodevolumeblockvolume_region" ON "bronze"."greennode_volume_block_volumes" ("region");
-- Create index "bronzegreennodevolumeblockvolume_status" to table: "greennode_volume_block_volumes"
CREATE INDEX "bronzegreennodevolumeblockvolume_status" ON "bronze"."greennode_volume_block_volumes" ("status");
-- Create "greennode_volume_snapshots" table
CREATE TABLE "bronze"."greennode_volume_snapshots" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "snapshot_id" character varying NOT NULL,
  "name" character varying NULL,
  "size" bigint NULL,
  "volume_size" bigint NULL,
  "status" character varying NULL,
  "created_at_api" character varying NULL,
  "bronze_green_node_volume_block_volume_snapshots" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "greennode_volume_snapshots_greennode_volume_block_volumes_snaps" FOREIGN KEY ("bronze_green_node_volume_block_volume_snapshots") REFERENCES "bronze"."greennode_volume_block_volumes" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
