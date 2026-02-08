-- Add new schema named "bronze"
CREATE SCHEMA "bronze";
-- Create "gcp_compute_target_tcp_proxies" table
CREATE TABLE "bronze"."gcp_compute_target_tcp_proxies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "service" character varying NULL,
  "proxy_bind" boolean NOT NULL DEFAULT false,
  "proxy_header" character varying NULL,
  "region" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargettcpproxy_collected_at" to table: "gcp_compute_target_tcp_proxies"
CREATE INDEX "bronzegcpcomputetargettcpproxy_collected_at" ON "bronze"."gcp_compute_target_tcp_proxies" ("collected_at");
-- Create index "bronzegcpcomputetargettcpproxy_project_id" to table: "gcp_compute_target_tcp_proxies"
CREATE INDEX "bronzegcpcomputetargettcpproxy_project_id" ON "bronze"."gcp_compute_target_tcp_proxies" ("project_id");
-- Create "gcp_iam_service_account_keys" table
CREATE TABLE "bronze"."gcp_iam_service_account_keys" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "service_account_email" character varying NOT NULL,
  "key_origin" character varying NULL,
  "key_type" character varying NULL,
  "key_algorithm" character varying NULL,
  "valid_after_time" timestamptz NULL,
  "valid_before_time" timestamptz NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpiamserviceaccountkey_collected_at" to table: "gcp_iam_service_account_keys"
CREATE INDEX "bronzegcpiamserviceaccountkey_collected_at" ON "bronze"."gcp_iam_service_account_keys" ("collected_at");
-- Create index "bronzegcpiamserviceaccountkey_project_id" to table: "gcp_iam_service_account_keys"
CREATE INDEX "bronzegcpiamserviceaccountkey_project_id" ON "bronze"."gcp_iam_service_account_keys" ("project_id");
-- Create index "bronzegcpiamserviceaccountkey_service_account_email" to table: "gcp_iam_service_account_keys"
CREATE INDEX "bronzegcpiamserviceaccountkey_service_account_email" ON "bronze"."gcp_iam_service_account_keys" ("service_account_email");
-- Create "gcp_compute_addresses" table
CREATE TABLE "bronze"."gcp_compute_addresses" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "address" character varying NULL,
  "address_type" character varying NULL,
  "ip_version" character varying NULL,
  "ipv6_endpoint_type" character varying NULL,
  "ip_collection" character varying NULL,
  "region" character varying NULL,
  "status" character varying NULL,
  "purpose" character varying NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "network_tier" character varying NULL,
  "prefix_length" integer NOT NULL DEFAULT 0,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "users_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeaddress_collected_at" to table: "gcp_compute_addresses"
CREATE INDEX "bronzegcpcomputeaddress_collected_at" ON "bronze"."gcp_compute_addresses" ("collected_at");
-- Create index "bronzegcpcomputeaddress_project_id" to table: "gcp_compute_addresses"
CREATE INDEX "bronzegcpcomputeaddress_project_id" ON "bronze"."gcp_compute_addresses" ("project_id");
-- Create index "bronzegcpcomputeaddress_status" to table: "gcp_compute_addresses"
CREATE INDEX "bronzegcpcomputeaddress_status" ON "bronze"."gcp_compute_addresses" ("status");
-- Create "gcp_compute_negs" table
CREATE TABLE "bronze"."gcp_compute_negs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "zone" character varying NULL,
  "network_endpoint_type" character varying NULL,
  "default_port" character varying NULL,
  "size" character varying NULL,
  "region" character varying NULL,
  "annotations_json" jsonb NULL,
  "app_engine_json" jsonb NULL,
  "cloud_function_json" jsonb NULL,
  "cloud_run_json" jsonb NULL,
  "psc_data_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeneg_collected_at" to table: "gcp_compute_negs"
CREATE INDEX "bronzegcpcomputeneg_collected_at" ON "bronze"."gcp_compute_negs" ("collected_at");
-- Create index "bronzegcpcomputeneg_project_id" to table: "gcp_compute_negs"
CREATE INDEX "bronzegcpcomputeneg_project_id" ON "bronze"."gcp_compute_negs" ("project_id");
-- Create index "bronzegcpcomputeneg_zone" to table: "gcp_compute_negs"
CREATE INDEX "bronzegcpcomputeneg_zone" ON "bronze"."gcp_compute_negs" ("zone");
-- Create "gcp_compute_health_checks" table
CREATE TABLE "bronze"."gcp_compute_health_checks" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "type" character varying NULL,
  "region" character varying NULL,
  "check_interval_sec" integer NULL,
  "timeout_sec" integer NULL,
  "healthy_threshold" integer NULL,
  "unhealthy_threshold" integer NULL,
  "tcp_health_check_json" jsonb NULL,
  "http_health_check_json" jsonb NULL,
  "https_health_check_json" jsonb NULL,
  "http2_health_check_json" jsonb NULL,
  "ssl_health_check_json" jsonb NULL,
  "grpc_health_check_json" jsonb NULL,
  "log_config_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputehealthcheck_collected_at" to table: "gcp_compute_health_checks"
CREATE INDEX "bronzegcpcomputehealthcheck_collected_at" ON "bronze"."gcp_compute_health_checks" ("collected_at");
-- Create index "bronzegcpcomputehealthcheck_project_id" to table: "gcp_compute_health_checks"
CREATE INDEX "bronzegcpcomputehealthcheck_project_id" ON "bronze"."gcp_compute_health_checks" ("project_id");
-- Create index "bronzegcpcomputehealthcheck_type" to table: "gcp_compute_health_checks"
CREATE INDEX "bronzegcpcomputehealthcheck_type" ON "bronze"."gcp_compute_health_checks" ("type");
-- Create "gcp_compute_target_http_proxies" table
CREATE TABLE "bronze"."gcp_compute_target_http_proxies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "fingerprint" character varying NULL,
  "url_map" character varying NULL,
  "proxy_bind" boolean NOT NULL DEFAULT false,
  "http_keep_alive_timeout_sec" integer NULL,
  "region" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargethttpproxy_collected_at" to table: "gcp_compute_target_http_proxies"
CREATE INDEX "bronzegcpcomputetargethttpproxy_collected_at" ON "bronze"."gcp_compute_target_http_proxies" ("collected_at");
-- Create index "bronzegcpcomputetargethttpproxy_project_id" to table: "gcp_compute_target_http_proxies"
CREATE INDEX "bronzegcpcomputetargethttpproxy_project_id" ON "bronze"."gcp_compute_target_http_proxies" ("project_id");
-- Create "gcp_compute_target_https_proxies" table
CREATE TABLE "bronze"."gcp_compute_target_https_proxies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "fingerprint" character varying NULL,
  "url_map" character varying NULL,
  "quic_override" character varying NULL,
  "server_tls_policy" character varying NULL,
  "authorization_policy" character varying NULL,
  "certificate_map" character varying NULL,
  "ssl_policy" character varying NULL,
  "tls_early_data" character varying NULL,
  "proxy_bind" boolean NOT NULL DEFAULT false,
  "http_keep_alive_timeout_sec" integer NULL,
  "ssl_certificates_json" jsonb NULL,
  "region" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargethttpsproxy_collected_at" to table: "gcp_compute_target_https_proxies"
CREATE INDEX "bronzegcpcomputetargethttpsproxy_collected_at" ON "bronze"."gcp_compute_target_https_proxies" ("collected_at");
-- Create index "bronzegcpcomputetargethttpsproxy_project_id" to table: "gcp_compute_target_https_proxies"
CREATE INDEX "bronzegcpcomputetargethttpsproxy_project_id" ON "bronze"."gcp_compute_target_https_proxies" ("project_id");
-- Create "gcp_compute_target_instances" table
CREATE TABLE "bronze"."gcp_compute_target_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "zone" character varying NULL,
  "instance" character varying NULL,
  "network" character varying NULL,
  "nat_policy" character varying NULL,
  "security_policy" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargetinstance_collected_at" to table: "gcp_compute_target_instances"
CREATE INDEX "bronzegcpcomputetargetinstance_collected_at" ON "bronze"."gcp_compute_target_instances" ("collected_at");
-- Create index "bronzegcpcomputetargetinstance_project_id" to table: "gcp_compute_target_instances"
CREATE INDEX "bronzegcpcomputetargetinstance_project_id" ON "bronze"."gcp_compute_target_instances" ("project_id");
-- Create "gcp_compute_target_ssl_proxies" table
CREATE TABLE "bronze"."gcp_compute_target_ssl_proxies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "service" character varying NULL,
  "proxy_header" character varying NULL,
  "certificate_map" character varying NULL,
  "ssl_policy" character varying NULL,
  "ssl_certificates_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargetsslproxy_collected_at" to table: "gcp_compute_target_ssl_proxies"
CREATE INDEX "bronzegcpcomputetargetsslproxy_collected_at" ON "bronze"."gcp_compute_target_ssl_proxies" ("collected_at");
-- Create index "bronzegcpcomputetargetsslproxy_project_id" to table: "gcp_compute_target_ssl_proxies"
CREATE INDEX "bronzegcpcomputetargetsslproxy_project_id" ON "bronze"."gcp_compute_target_ssl_proxies" ("project_id");
-- Create "gcp_compute_neg_endpoints" table
CREATE TABLE "bronze"."gcp_compute_neg_endpoints" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "instance" character varying NULL,
  "ip_address" character varying NULL,
  "ipv6_address" character varying NULL,
  "port" character varying NULL,
  "fqdn" character varying NULL,
  "annotations_json" jsonb NULL,
  "neg_name" character varying NOT NULL,
  "zone" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputenegendpoint_collected_at" to table: "gcp_compute_neg_endpoints"
CREATE INDEX "bronzegcpcomputenegendpoint_collected_at" ON "bronze"."gcp_compute_neg_endpoints" ("collected_at");
-- Create index "bronzegcpcomputenegendpoint_neg_name" to table: "gcp_compute_neg_endpoints"
CREATE INDEX "bronzegcpcomputenegendpoint_neg_name" ON "bronze"."gcp_compute_neg_endpoints" ("neg_name");
-- Create index "bronzegcpcomputenegendpoint_project_id" to table: "gcp_compute_neg_endpoints"
CREATE INDEX "bronzegcpcomputenegendpoint_project_id" ON "bronze"."gcp_compute_neg_endpoints" ("project_id");
-- Create "gcp_compute_url_maps" table
CREATE TABLE "bronze"."gcp_compute_url_maps" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "fingerprint" character varying NULL,
  "default_service" character varying NULL,
  "region" character varying NULL,
  "host_rules_json" jsonb NULL,
  "path_matchers_json" jsonb NULL,
  "tests_json" jsonb NULL,
  "default_route_action_json" jsonb NULL,
  "default_url_redirect_json" jsonb NULL,
  "header_action_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeurlmap_collected_at" to table: "gcp_compute_url_maps"
CREATE INDEX "bronzegcpcomputeurlmap_collected_at" ON "bronze"."gcp_compute_url_maps" ("collected_at");
-- Create index "bronzegcpcomputeurlmap_project_id" to table: "gcp_compute_url_maps"
CREATE INDEX "bronzegcpcomputeurlmap_project_id" ON "bronze"."gcp_compute_url_maps" ("project_id");
-- Create "gcp_iam_service_accounts" table
CREATE TABLE "bronze"."gcp_iam_service_accounts" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "email" character varying NOT NULL,
  "display_name" character varying NULL,
  "description" character varying NULL,
  "oauth2_client_id" character varying NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpiamserviceaccount_collected_at" to table: "gcp_iam_service_accounts"
CREATE INDEX "bronzegcpiamserviceaccount_collected_at" ON "bronze"."gcp_iam_service_accounts" ("collected_at");
-- Create index "bronzegcpiamserviceaccount_email" to table: "gcp_iam_service_accounts"
CREATE INDEX "bronzegcpiamserviceaccount_email" ON "bronze"."gcp_iam_service_accounts" ("email");
-- Create index "bronzegcpiamserviceaccount_project_id" to table: "gcp_iam_service_accounts"
CREATE INDEX "bronzegcpiamserviceaccount_project_id" ON "bronze"."gcp_iam_service_accounts" ("project_id");
-- Create "gcp_vpc_access_connectors" table
CREATE TABLE "bronze"."gcp_vpc_access_connectors" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "network" character varying NULL,
  "ip_cidr_range" character varying NULL,
  "state" character varying NULL,
  "min_throughput" integer NULL,
  "max_throughput" integer NULL,
  "min_instances" integer NULL,
  "max_instances" integer NULL,
  "machine_type" character varying NULL,
  "region" character varying NULL,
  "subnet_json" jsonb NULL,
  "connected_projects_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpvpcaccessconnector_collected_at" to table: "gcp_vpc_access_connectors"
CREATE INDEX "bronzegcpvpcaccessconnector_collected_at" ON "bronze"."gcp_vpc_access_connectors" ("collected_at");
-- Create index "bronzegcpvpcaccessconnector_project_id" to table: "gcp_vpc_access_connectors"
CREATE INDEX "bronzegcpvpcaccessconnector_project_id" ON "bronze"."gcp_vpc_access_connectors" ("project_id");
-- Create "gcp_compute_target_pools" table
CREATE TABLE "bronze"."gcp_compute_target_pools" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "session_affinity" character varying NULL,
  "backup_pool" character varying NULL,
  "failover_ratio" real NULL,
  "security_policy" character varying NULL,
  "region" character varying NULL,
  "health_checks_json" jsonb NULL,
  "instances_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputetargetpool_collected_at" to table: "gcp_compute_target_pools"
CREATE INDEX "bronzegcpcomputetargetpool_collected_at" ON "bronze"."gcp_compute_target_pools" ("collected_at");
-- Create index "bronzegcpcomputetargetpool_project_id" to table: "gcp_compute_target_pools"
CREATE INDEX "bronzegcpcomputetargetpool_project_id" ON "bronze"."gcp_compute_target_pools" ("project_id");
-- Create index "bronzegcpcomputetargetpool_region" to table: "gcp_compute_target_pools"
CREATE INDEX "bronzegcpcomputetargetpool_region" ON "bronze"."gcp_compute_target_pools" ("region");
-- Create "gcp_compute_address_labels" table
CREATE TABLE "bronze"."gcp_compute_address_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_address_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_address_labels_gcp_compute_addresses_labels" FOREIGN KEY ("bronze_gcp_compute_address_labels") REFERENCES "bronze"."gcp_compute_addresses" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_backend_services" table
CREATE TABLE "bronze"."gcp_compute_backend_services" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "self_link" character varying NULL,
  "fingerprint" character varying NULL,
  "load_balancing_scheme" character varying NULL,
  "protocol" character varying NULL,
  "port_name" character varying NULL,
  "port" character varying NULL,
  "timeout_sec" character varying NULL,
  "region" character varying NULL,
  "network" character varying NULL,
  "security_policy" character varying NULL,
  "edge_security_policy" character varying NULL,
  "session_affinity" character varying NULL,
  "affinity_cookie_ttl_sec" character varying NULL,
  "locality_lb_policy" character varying NULL,
  "compression_mode" character varying NULL,
  "service_lb_policy" character varying NULL,
  "enable_cdn" boolean NOT NULL DEFAULT false,
  "health_checks_json" jsonb NULL,
  "locality_lb_policies_json" jsonb NULL,
  "used_by_json" jsonb NULL,
  "custom_request_headers_json" jsonb NULL,
  "custom_response_headers_json" jsonb NULL,
  "cdn_policy_json" jsonb NULL,
  "circuit_breakers_json" jsonb NULL,
  "connection_draining_json" jsonb NULL,
  "connection_tracking_policy_json" jsonb NULL,
  "consistent_hash_json" jsonb NULL,
  "failover_policy_json" jsonb NULL,
  "iap_json" jsonb NULL,
  "log_config_json" jsonb NULL,
  "max_stream_duration_json" jsonb NULL,
  "outlier_detection_json" jsonb NULL,
  "security_settings_json" jsonb NULL,
  "subsetting_json" jsonb NULL,
  "service_bindings_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputebackendservice_collected_at" to table: "gcp_compute_backend_services"
CREATE INDEX "bronzegcpcomputebackendservice_collected_at" ON "bronze"."gcp_compute_backend_services" ("collected_at");
-- Create index "bronzegcpcomputebackendservice_project_id" to table: "gcp_compute_backend_services"
CREATE INDEX "bronzegcpcomputebackendservice_project_id" ON "bronze"."gcp_compute_backend_services" ("project_id");
-- Create "gcp_compute_backend_service_backends" table
CREATE TABLE "bronze"."gcp_compute_backend_service_backends" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "group" character varying NOT NULL,
  "balancing_mode" character varying NULL,
  "capacity_scaler" character varying NULL,
  "description" character varying NULL,
  "failover" boolean NOT NULL DEFAULT false,
  "max_connections" character varying NULL,
  "max_connections_per_endpoint" character varying NULL,
  "max_connections_per_instance" character varying NULL,
  "max_rate" character varying NULL,
  "max_rate_per_endpoint" character varying NULL,
  "max_rate_per_instance" character varying NULL,
  "max_utilization" character varying NULL,
  "preference" character varying NULL,
  "bronze_gcp_compute_backend_service_backends" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_backend_service_backends_gcp_compute_backend_servic" FOREIGN KEY ("bronze_gcp_compute_backend_service_backends") REFERENCES "bronze"."gcp_compute_backend_services" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_disks" table
CREATE TABLE "bronze"."gcp_compute_disks" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "zone" character varying NULL,
  "region" character varying NULL,
  "type" character varying NULL,
  "status" character varying NULL,
  "size_gb" bigint NULL,
  "architecture" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "last_attach_timestamp" character varying NULL,
  "last_detach_timestamp" character varying NULL,
  "source_image" character varying NULL,
  "source_image_id" character varying NULL,
  "source_snapshot" character varying NULL,
  "source_snapshot_id" character varying NULL,
  "source_disk" character varying NULL,
  "source_disk_id" character varying NULL,
  "provisioned_iops" bigint NULL,
  "provisioned_throughput" bigint NULL,
  "physical_block_size_bytes" bigint NULL,
  "enable_confidential_compute" boolean NOT NULL DEFAULT false,
  "disk_encryption_key_json" jsonb NULL,
  "users_json" jsonb NULL,
  "replica_zones_json" jsonb NULL,
  "resource_policies_json" jsonb NULL,
  "guest_os_features_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputedisk_collected_at" to table: "gcp_compute_disks"
CREATE INDEX "bronzegcpcomputedisk_collected_at" ON "bronze"."gcp_compute_disks" ("collected_at");
-- Create index "bronzegcpcomputedisk_project_id" to table: "gcp_compute_disks"
CREATE INDEX "bronzegcpcomputedisk_project_id" ON "bronze"."gcp_compute_disks" ("project_id");
-- Create index "bronzegcpcomputedisk_status" to table: "gcp_compute_disks"
CREATE INDEX "bronzegcpcomputedisk_status" ON "bronze"."gcp_compute_disks" ("status");
-- Create "gcp_compute_disk_labels" table
CREATE TABLE "bronze"."gcp_compute_disk_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_disk_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_disk_labels_gcp_compute_disks_labels" FOREIGN KEY ("bronze_gcp_compute_disk_labels") REFERENCES "bronze"."gcp_compute_disks" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_disk_licenses" table
CREATE TABLE "bronze"."gcp_compute_disk_licenses" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "license" character varying NOT NULL,
  "bronze_gcp_compute_disk_licenses" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_disk_licenses_gcp_compute_disks_licenses" FOREIGN KEY ("bronze_gcp_compute_disk_licenses") REFERENCES "bronze"."gcp_compute_disks" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_forwarding_rules" table
CREATE TABLE "bronze"."gcp_compute_forwarding_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "ip_address" character varying NULL,
  "ip_protocol" character varying NULL,
  "all_ports" boolean NOT NULL DEFAULT false,
  "allow_global_access" boolean NOT NULL DEFAULT false,
  "allow_psc_global_access" boolean NOT NULL DEFAULT false,
  "backend_service" character varying NULL,
  "base_forwarding_rule" character varying NULL,
  "creation_timestamp" character varying NULL,
  "external_managed_backend_bucket_migration_state" character varying NULL,
  "external_managed_backend_bucket_migration_testing_percentage" real NULL,
  "fingerprint" character varying NULL,
  "ip_collection" character varying NULL,
  "ip_version" character varying NULL,
  "is_mirroring_collector" boolean NOT NULL DEFAULT false,
  "label_fingerprint" character varying NULL,
  "load_balancing_scheme" character varying NULL,
  "network" character varying NULL,
  "network_tier" character varying NULL,
  "no_automate_dns_zone" boolean NOT NULL DEFAULT false,
  "port_range" character varying NULL,
  "psc_connection_id" character varying NULL,
  "psc_connection_status" character varying NULL,
  "region" character varying NULL,
  "self_link" character varying NULL,
  "self_link_with_id" character varying NULL,
  "service_label" character varying NULL,
  "service_name" character varying NULL,
  "subnetwork" character varying NULL,
  "target" character varying NULL,
  "ports_json" jsonb NULL,
  "source_ip_ranges_json" jsonb NULL,
  "metadata_filters_json" jsonb NULL,
  "service_directory_registrations_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeforwardingrule_collected_at" to table: "gcp_compute_forwarding_rules"
CREATE INDEX "bronzegcpcomputeforwardingrule_collected_at" ON "bronze"."gcp_compute_forwarding_rules" ("collected_at");
-- Create index "bronzegcpcomputeforwardingrule_project_id" to table: "gcp_compute_forwarding_rules"
CREATE INDEX "bronzegcpcomputeforwardingrule_project_id" ON "bronze"."gcp_compute_forwarding_rules" ("project_id");
-- Create index "bronzegcpcomputeforwardingrule_region" to table: "gcp_compute_forwarding_rules"
CREATE INDEX "bronzegcpcomputeforwardingrule_region" ON "bronze"."gcp_compute_forwarding_rules" ("region");
-- Create "gcp_compute_forwarding_rule_labels" table
CREATE TABLE "bronze"."gcp_compute_forwarding_rule_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_forwarding_rule_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_forwarding_rule_labels_gcp_compute_forwarding_rules" FOREIGN KEY ("bronze_gcp_compute_forwarding_rule_labels") REFERENCES "bronze"."gcp_compute_forwarding_rules" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_global_addresses" table
CREATE TABLE "bronze"."gcp_compute_global_addresses" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "address" character varying NULL,
  "address_type" character varying NULL,
  "ip_version" character varying NULL,
  "ipv6_endpoint_type" character varying NULL,
  "ip_collection" character varying NULL,
  "region" character varying NULL,
  "status" character varying NULL,
  "purpose" character varying NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "network_tier" character varying NULL,
  "prefix_length" integer NOT NULL DEFAULT 0,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "users_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeglobaladdress_collected_at" to table: "gcp_compute_global_addresses"
CREATE INDEX "bronzegcpcomputeglobaladdress_collected_at" ON "bronze"."gcp_compute_global_addresses" ("collected_at");
-- Create index "bronzegcpcomputeglobaladdress_project_id" to table: "gcp_compute_global_addresses"
CREATE INDEX "bronzegcpcomputeglobaladdress_project_id" ON "bronze"."gcp_compute_global_addresses" ("project_id");
-- Create index "bronzegcpcomputeglobaladdress_status" to table: "gcp_compute_global_addresses"
CREATE INDEX "bronzegcpcomputeglobaladdress_status" ON "bronze"."gcp_compute_global_addresses" ("status");
-- Create "gcp_compute_global_address_labels" table
CREATE TABLE "bronze"."gcp_compute_global_address_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_global_address_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_global_address_labels_gcp_compute_global_addresses_" FOREIGN KEY ("bronze_gcp_compute_global_address_labels") REFERENCES "bronze"."gcp_compute_global_addresses" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_global_forwarding_rules" table
CREATE TABLE "bronze"."gcp_compute_global_forwarding_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "ip_address" character varying NULL,
  "ip_protocol" character varying NULL,
  "all_ports" boolean NOT NULL DEFAULT false,
  "allow_global_access" boolean NOT NULL DEFAULT false,
  "allow_psc_global_access" boolean NOT NULL DEFAULT false,
  "backend_service" character varying NULL,
  "base_forwarding_rule" character varying NULL,
  "creation_timestamp" character varying NULL,
  "external_managed_backend_bucket_migration_state" character varying NULL,
  "external_managed_backend_bucket_migration_testing_percentage" real NULL,
  "fingerprint" character varying NULL,
  "ip_collection" character varying NULL,
  "ip_version" character varying NULL,
  "is_mirroring_collector" boolean NOT NULL DEFAULT false,
  "label_fingerprint" character varying NULL,
  "load_balancing_scheme" character varying NULL,
  "network" character varying NULL,
  "network_tier" character varying NULL,
  "no_automate_dns_zone" boolean NOT NULL DEFAULT false,
  "port_range" character varying NULL,
  "psc_connection_id" character varying NULL,
  "psc_connection_status" character varying NULL,
  "region" character varying NULL,
  "self_link" character varying NULL,
  "self_link_with_id" character varying NULL,
  "service_label" character varying NULL,
  "service_name" character varying NULL,
  "subnetwork" character varying NULL,
  "target" character varying NULL,
  "ports_json" jsonb NULL,
  "source_ip_ranges_json" jsonb NULL,
  "metadata_filters_json" jsonb NULL,
  "service_directory_registrations_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeglobalforwardingrule_collected_at" to table: "gcp_compute_global_forwarding_rules"
CREATE INDEX "bronzegcpcomputeglobalforwardingrule_collected_at" ON "bronze"."gcp_compute_global_forwarding_rules" ("collected_at");
-- Create index "bronzegcpcomputeglobalforwardingrule_project_id" to table: "gcp_compute_global_forwarding_rules"
CREATE INDEX "bronzegcpcomputeglobalforwardingrule_project_id" ON "bronze"."gcp_compute_global_forwarding_rules" ("project_id");
-- Create "gcp_compute_global_forwarding_rule_labels" table
CREATE TABLE "bronze"."gcp_compute_global_forwarding_rule_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_global_forwarding_rule_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_global_forwarding_rule_labels_gcp_compute_global_fo" FOREIGN KEY ("bronze_gcp_compute_global_forwarding_rule_labels") REFERENCES "bronze"."gcp_compute_global_forwarding_rules" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_images" table
CREATE TABLE "bronze"."gcp_compute_images" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "architecture" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "family" character varying NULL,
  "source_disk" character varying NULL,
  "source_disk_id" character varying NULL,
  "source_image" character varying NULL,
  "source_image_id" character varying NULL,
  "source_snapshot" character varying NULL,
  "source_snapshot_id" character varying NULL,
  "source_type" character varying NULL,
  "disk_size_gb" bigint NULL,
  "archive_size_bytes" bigint NULL,
  "satisfies_pzi" boolean NOT NULL DEFAULT false,
  "satisfies_pzs" boolean NOT NULL DEFAULT false,
  "enable_confidential_compute" boolean NOT NULL DEFAULT false,
  "image_encryption_key_json" jsonb NULL,
  "source_disk_encryption_key_json" jsonb NULL,
  "source_image_encryption_key_json" jsonb NULL,
  "source_snapshot_encryption_key_json" jsonb NULL,
  "deprecated_json" jsonb NULL,
  "guest_os_features_json" jsonb NULL,
  "shielded_instance_initial_state_json" jsonb NULL,
  "raw_disk_json" jsonb NULL,
  "storage_locations_json" jsonb NULL,
  "license_codes_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeimage_collected_at" to table: "gcp_compute_images"
CREATE INDEX "bronzegcpcomputeimage_collected_at" ON "bronze"."gcp_compute_images" ("collected_at");
-- Create index "bronzegcpcomputeimage_project_id" to table: "gcp_compute_images"
CREATE INDEX "bronzegcpcomputeimage_project_id" ON "bronze"."gcp_compute_images" ("project_id");
-- Create index "bronzegcpcomputeimage_status" to table: "gcp_compute_images"
CREATE INDEX "bronzegcpcomputeimage_status" ON "bronze"."gcp_compute_images" ("status");
-- Create "gcp_compute_image_labels" table
CREATE TABLE "bronze"."gcp_compute_image_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_image_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_image_labels_gcp_compute_images_labels" FOREIGN KEY ("bronze_gcp_compute_image_labels") REFERENCES "bronze"."gcp_compute_images" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_image_licenses" table
CREATE TABLE "bronze"."gcp_compute_image_licenses" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "license" character varying NOT NULL,
  "bronze_gcp_compute_image_licenses" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_image_licenses_gcp_compute_images_licenses" FOREIGN KEY ("bronze_gcp_compute_image_licenses") REFERENCES "bronze"."gcp_compute_images" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instances" table
CREATE TABLE "bronze"."gcp_compute_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "zone" character varying NULL,
  "machine_type" character varying NULL,
  "status" character varying NULL,
  "status_message" character varying NULL,
  "cpu_platform" character varying NULL,
  "hostname" character varying NULL,
  "description" character varying NULL,
  "creation_timestamp" character varying NULL,
  "last_start_timestamp" character varying NULL,
  "last_stop_timestamp" character varying NULL,
  "last_suspended_timestamp" character varying NULL,
  "deletion_protection" boolean NOT NULL DEFAULT false,
  "can_ip_forward" boolean NOT NULL DEFAULT false,
  "self_link" character varying NULL,
  "scheduling_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeinstance_collected_at" to table: "gcp_compute_instances"
CREATE INDEX "bronzegcpcomputeinstance_collected_at" ON "bronze"."gcp_compute_instances" ("collected_at");
-- Create index "bronzegcpcomputeinstance_project_id" to table: "gcp_compute_instances"
CREATE INDEX "bronzegcpcomputeinstance_project_id" ON "bronze"."gcp_compute_instances" ("project_id");
-- Create index "bronzegcpcomputeinstance_status" to table: "gcp_compute_instances"
CREATE INDEX "bronzegcpcomputeinstance_status" ON "bronze"."gcp_compute_instances" ("status");
-- Create "gcp_compute_instance_disks" table
CREATE TABLE "bronze"."gcp_compute_instance_disks" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "source" character varying NULL,
  "device_name" character varying NULL,
  "index" bigint NULL,
  "boot" boolean NOT NULL DEFAULT false,
  "auto_delete" boolean NOT NULL DEFAULT false,
  "mode" character varying NULL,
  "interface" character varying NULL,
  "type" character varying NULL,
  "disk_size_gb" bigint NULL,
  "disk_encryption_key_json" jsonb NULL,
  "initialize_params_json" jsonb NULL,
  "bronze_gcp_compute_instance_disks" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_disks_gcp_compute_instances_disks" FOREIGN KEY ("bronze_gcp_compute_instance_disks") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_disk_licenses" table
CREATE TABLE "bronze"."gcp_compute_instance_disk_licenses" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "license" character varying NOT NULL,
  "bronze_gcp_compute_instance_disk_licenses" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_disk_licenses_gcp_compute_instance_disks_l" FOREIGN KEY ("bronze_gcp_compute_instance_disk_licenses") REFERENCES "bronze"."gcp_compute_instance_disks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_groups" table
CREATE TABLE "bronze"."gcp_compute_instance_groups" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "zone" character varying NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "size" integer NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "fingerprint" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeinstancegroup_collected_at" to table: "gcp_compute_instance_groups"
CREATE INDEX "bronzegcpcomputeinstancegroup_collected_at" ON "bronze"."gcp_compute_instance_groups" ("collected_at");
-- Create index "bronzegcpcomputeinstancegroup_project_id" to table: "gcp_compute_instance_groups"
CREATE INDEX "bronzegcpcomputeinstancegroup_project_id" ON "bronze"."gcp_compute_instance_groups" ("project_id");
-- Create "gcp_compute_instance_group_members" table
CREATE TABLE "bronze"."gcp_compute_instance_group_members" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "instance_url" character varying NOT NULL,
  "instance_name" character varying NULL,
  "status" character varying NULL,
  "bronze_gcp_compute_instance_group_members" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_group_members_gcp_compute_instance_groups_" FOREIGN KEY ("bronze_gcp_compute_instance_group_members") REFERENCES "bronze"."gcp_compute_instance_groups" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_group_named_ports" table
CREATE TABLE "bronze"."gcp_compute_instance_group_named_ports" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "name" character varying NOT NULL,
  "port" integer NOT NULL,
  "bronze_gcp_compute_instance_group_named_ports" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_group_named_ports_gcp_compute_instance_gro" FOREIGN KEY ("bronze_gcp_compute_instance_group_named_ports") REFERENCES "bronze"."gcp_compute_instance_groups" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_labels" table
CREATE TABLE "bronze"."gcp_compute_instance_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NULL,
  "bronze_gcp_compute_instance_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_labels_gcp_compute_instances_labels" FOREIGN KEY ("bronze_gcp_compute_instance_labels") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_metadata" table
CREATE TABLE "bronze"."gcp_compute_instance_metadata" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NULL,
  "bronze_gcp_compute_instance_metadata" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_metadata_gcp_compute_instances_metadata" FOREIGN KEY ("bronze_gcp_compute_instance_metadata") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_nics" table
CREATE TABLE "bronze"."gcp_compute_instance_nics" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "name" character varying NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "network_ip" character varying NULL,
  "stack_type" character varying NULL,
  "nic_type" character varying NULL,
  "bronze_gcp_compute_instance_nics" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_nics_gcp_compute_instances_nics" FOREIGN KEY ("bronze_gcp_compute_instance_nics") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_nic_access_configs" table
CREATE TABLE "bronze"."gcp_compute_instance_nic_access_configs" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "type" character varying NULL,
  "name" character varying NULL,
  "nat_ip" character varying NULL,
  "network_tier" character varying NULL,
  "bronze_gcp_compute_instance_nic_access_configs" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_nic_access_configs_gcp_compute_instance_ni" FOREIGN KEY ("bronze_gcp_compute_instance_nic_access_configs") REFERENCES "bronze"."gcp_compute_instance_nics" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_nic_alias_ranges" table
CREATE TABLE "bronze"."gcp_compute_instance_nic_alias_ranges" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "ip_cidr_range" character varying NULL,
  "subnetwork_range_name" character varying NULL,
  "bronze_gcp_compute_instance_nic_alias_ip_ranges" bigint NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_nic_alias_ranges_gcp_compute_instance_nics" FOREIGN KEY ("bronze_gcp_compute_instance_nic_alias_ip_ranges") REFERENCES "bronze"."gcp_compute_instance_nics" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_service_accounts" table
CREATE TABLE "bronze"."gcp_compute_instance_service_accounts" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "email" character varying NOT NULL,
  "scopes_json" jsonb NULL,
  "bronze_gcp_compute_instance_service_accounts" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_service_accounts_gcp_compute_instances_ser" FOREIGN KEY ("bronze_gcp_compute_instance_service_accounts") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_instance_tags" table
CREATE TABLE "bronze"."gcp_compute_instance_tags" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "tag" character varying NOT NULL,
  "bronze_gcp_compute_instance_tags" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_instance_tags_gcp_compute_instances_tags" FOREIGN KEY ("bronze_gcp_compute_instance_tags") REFERENCES "bronze"."gcp_compute_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_networks" table
CREATE TABLE "bronze"."gcp_compute_networks" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "auto_create_subnetworks" boolean NOT NULL DEFAULT false,
  "mtu" bigint NOT NULL DEFAULT 0,
  "routing_mode" character varying NULL,
  "network_firewall_policy_enforcement_order" character varying NULL,
  "enable_ula_internal_ipv6" boolean NOT NULL DEFAULT false,
  "internal_ipv6_range" character varying NULL,
  "gateway_ipv4" character varying NULL,
  "subnetworks_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputenetwork_collected_at" to table: "gcp_compute_networks"
CREATE INDEX "bronzegcpcomputenetwork_collected_at" ON "bronze"."gcp_compute_networks" ("collected_at");
-- Create index "bronzegcpcomputenetwork_project_id" to table: "gcp_compute_networks"
CREATE INDEX "bronzegcpcomputenetwork_project_id" ON "bronze"."gcp_compute_networks" ("project_id");
-- Create "gcp_compute_network_peerings" table
CREATE TABLE "bronze"."gcp_compute_network_peerings" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "name" character varying NOT NULL,
  "network" character varying NULL,
  "state" character varying NULL,
  "state_details" character varying NULL,
  "export_custom_routes" boolean NOT NULL DEFAULT false,
  "import_custom_routes" boolean NOT NULL DEFAULT false,
  "export_subnet_routes_with_public_ip" boolean NOT NULL DEFAULT false,
  "import_subnet_routes_with_public_ip" boolean NOT NULL DEFAULT false,
  "exchange_subnet_routes" boolean NOT NULL DEFAULT false,
  "stack_type" character varying NULL,
  "peer_mtu" bigint NOT NULL DEFAULT 0,
  "auto_create_routes" boolean NOT NULL DEFAULT false,
  "bronze_gcp_compute_network_peerings" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_network_peerings_gcp_compute_networks_peerings" FOREIGN KEY ("bronze_gcp_compute_network_peerings") REFERENCES "bronze"."gcp_compute_networks" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_snapshots" table
CREATE TABLE "bronze"."gcp_compute_snapshots" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "disk_size_gb" bigint NULL,
  "storage_bytes" bigint NULL,
  "storage_bytes_status" character varying NULL,
  "download_bytes" bigint NULL,
  "snapshot_type" character varying NULL,
  "architecture" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "source_disk" character varying NULL,
  "source_disk_id" character varying NULL,
  "source_disk_for_recovery_checkpoint" character varying NULL,
  "auto_created" boolean NOT NULL DEFAULT false,
  "satisfies_pzi" boolean NOT NULL DEFAULT false,
  "satisfies_pzs" boolean NOT NULL DEFAULT false,
  "enable_confidential_compute" boolean NOT NULL DEFAULT false,
  "snapshot_encryption_key_json" jsonb NULL,
  "source_disk_encryption_key_json" jsonb NULL,
  "guest_os_features_json" jsonb NULL,
  "storage_locations_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputesnapshot_collected_at" to table: "gcp_compute_snapshots"
CREATE INDEX "bronzegcpcomputesnapshot_collected_at" ON "bronze"."gcp_compute_snapshots" ("collected_at");
-- Create index "bronzegcpcomputesnapshot_project_id" to table: "gcp_compute_snapshots"
CREATE INDEX "bronzegcpcomputesnapshot_project_id" ON "bronze"."gcp_compute_snapshots" ("project_id");
-- Create index "bronzegcpcomputesnapshot_status" to table: "gcp_compute_snapshots"
CREATE INDEX "bronzegcpcomputesnapshot_status" ON "bronze"."gcp_compute_snapshots" ("status");
-- Create "gcp_compute_snapshot_labels" table
CREATE TABLE "bronze"."gcp_compute_snapshot_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_compute_snapshot_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_snapshot_labels_gcp_compute_snapshots_labels" FOREIGN KEY ("bronze_gcp_compute_snapshot_labels") REFERENCES "bronze"."gcp_compute_snapshots" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_snapshot_licenses" table
CREATE TABLE "bronze"."gcp_compute_snapshot_licenses" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "license" character varying NOT NULL,
  "bronze_gcp_compute_snapshot_licenses" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_snapshot_licenses_gcp_compute_snapshots_licenses" FOREIGN KEY ("bronze_gcp_compute_snapshot_licenses") REFERENCES "bronze"."gcp_compute_snapshots" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_subnetworks" table
CREATE TABLE "bronze"."gcp_compute_subnetworks" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "network" character varying NOT NULL,
  "region" character varying NOT NULL,
  "ip_cidr_range" character varying NOT NULL,
  "gateway_address" character varying NULL,
  "purpose" character varying NULL,
  "role" character varying NULL,
  "private_ip_google_access" boolean NOT NULL DEFAULT false,
  "private_ipv6_google_access" character varying NULL,
  "stack_type" character varying NULL,
  "ipv6_access_type" character varying NULL,
  "internal_ipv6_prefix" character varying NULL,
  "external_ipv6_prefix" character varying NULL,
  "log_config_json" jsonb NULL,
  "fingerprint" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputesubnetwork_collected_at" to table: "gcp_compute_subnetworks"
CREATE INDEX "bronzegcpcomputesubnetwork_collected_at" ON "bronze"."gcp_compute_subnetworks" ("collected_at");
-- Create index "bronzegcpcomputesubnetwork_project_id" to table: "gcp_compute_subnetworks"
CREATE INDEX "bronzegcpcomputesubnetwork_project_id" ON "bronze"."gcp_compute_subnetworks" ("project_id");
-- Create "gcp_compute_subnetwork_secondary_ranges" table
CREATE TABLE "bronze"."gcp_compute_subnetwork_secondary_ranges" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "range_name" character varying NOT NULL,
  "ip_cidr_range" character varying NOT NULL,
  "bronze_gcp_compute_subnetwork_secondary_ip_ranges" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_subnetwork_secondary_ranges_gcp_compute_subnetworks" FOREIGN KEY ("bronze_gcp_compute_subnetwork_secondary_ip_ranges") REFERENCES "bronze"."gcp_compute_subnetworks" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_target_vpn_gateways" table
CREATE TABLE "bronze"."gcp_compute_target_vpn_gateways" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "region" character varying NULL,
  "network" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "forwarding_rules_json" jsonb NULL,
  "tunnels_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpvpntargetgateway_collected_at" to table: "gcp_compute_target_vpn_gateways"
CREATE INDEX "bronzegcpvpntargetgateway_collected_at" ON "bronze"."gcp_compute_target_vpn_gateways" ("collected_at");
-- Create index "bronzegcpvpntargetgateway_project_id" to table: "gcp_compute_target_vpn_gateways"
CREATE INDEX "bronzegcpvpntargetgateway_project_id" ON "bronze"."gcp_compute_target_vpn_gateways" ("project_id");
-- Create index "bronzegcpvpntargetgateway_status" to table: "gcp_compute_target_vpn_gateways"
CREATE INDEX "bronzegcpvpntargetgateway_status" ON "bronze"."gcp_compute_target_vpn_gateways" ("status");
-- Create "gcp_compute_target_vpn_gateway_labels" table
CREATE TABLE "bronze"."gcp_compute_target_vpn_gateway_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcpvpn_target_gateway_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_target_vpn_gateway_labels_gcp_compute_target_vpn_ga" FOREIGN KEY ("bronze_gcpvpn_target_gateway_labels") REFERENCES "bronze"."gcp_compute_target_vpn_gateways" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_vpn_gateways" table
CREATE TABLE "bronze"."gcp_compute_vpn_gateways" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "region" character varying NULL,
  "network" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "gateway_ip_version" character varying NULL,
  "stack_type" character varying NULL,
  "vpn_interfaces_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpvpngateway_collected_at" to table: "gcp_compute_vpn_gateways"
CREATE INDEX "bronzegcpvpngateway_collected_at" ON "bronze"."gcp_compute_vpn_gateways" ("collected_at");
-- Create index "bronzegcpvpngateway_project_id" to table: "gcp_compute_vpn_gateways"
CREATE INDEX "bronzegcpvpngateway_project_id" ON "bronze"."gcp_compute_vpn_gateways" ("project_id");
-- Create "gcp_compute_vpn_gateway_labels" table
CREATE TABLE "bronze"."gcp_compute_vpn_gateway_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcpvpn_gateway_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_vpn_gateway_labels_gcp_compute_vpn_gateways_labels" FOREIGN KEY ("bronze_gcpvpn_gateway_labels") REFERENCES "bronze"."gcp_compute_vpn_gateways" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_vpn_tunnels" table
CREATE TABLE "bronze"."gcp_compute_vpn_tunnels" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "status" character varying NULL,
  "detailed_status" character varying NULL,
  "region" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "label_fingerprint" character varying NULL,
  "ike_version" integer NULL,
  "peer_ip" character varying NULL,
  "peer_external_gateway" character varying NULL,
  "peer_external_gateway_interface" integer NULL,
  "peer_gcp_gateway" character varying NULL,
  "router" character varying NULL,
  "shared_secret_hash" character varying NULL,
  "vpn_gateway" character varying NULL,
  "target_vpn_gateway" character varying NULL,
  "vpn_gateway_interface" integer NULL,
  "local_traffic_selector_json" jsonb NULL,
  "remote_traffic_selector_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpvpntunnel_collected_at" to table: "gcp_compute_vpn_tunnels"
CREATE INDEX "bronzegcpvpntunnel_collected_at" ON "bronze"."gcp_compute_vpn_tunnels" ("collected_at");
-- Create index "bronzegcpvpntunnel_project_id" to table: "gcp_compute_vpn_tunnels"
CREATE INDEX "bronzegcpvpntunnel_project_id" ON "bronze"."gcp_compute_vpn_tunnels" ("project_id");
-- Create index "bronzegcpvpntunnel_status" to table: "gcp_compute_vpn_tunnels"
CREATE INDEX "bronzegcpvpntunnel_status" ON "bronze"."gcp_compute_vpn_tunnels" ("status");
-- Create "gcp_compute_vpn_tunnel_labels" table
CREATE TABLE "bronze"."gcp_compute_vpn_tunnel_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcpvpn_tunnel_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_vpn_tunnel_labels_gcp_compute_vpn_tunnels_labels" FOREIGN KEY ("bronze_gcpvpn_tunnel_labels") REFERENCES "bronze"."gcp_compute_vpn_tunnels" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_container_clusters" table
CREATE TABLE "bronze"."gcp_container_clusters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "location" character varying NULL,
  "zone" character varying NULL,
  "description" character varying NULL,
  "initial_cluster_version" character varying NULL,
  "current_master_version" character varying NULL,
  "current_node_version" character varying NULL,
  "status" character varying NULL,
  "status_message" character varying NULL,
  "current_node_count" integer NULL,
  "network" character varying NULL,
  "subnetwork" character varying NULL,
  "cluster_ipv4_cidr" character varying NULL,
  "services_ipv4_cidr" character varying NULL,
  "node_ipv4_cidr_size" integer NULL,
  "endpoint" character varying NULL,
  "self_link" character varying NULL,
  "create_time" character varying NULL,
  "expire_time" character varying NULL,
  "etag" character varying NULL,
  "label_fingerprint" character varying NULL,
  "logging_service" character varying NULL,
  "monitoring_service" character varying NULL,
  "enable_kubernetes_alpha" boolean NOT NULL DEFAULT false,
  "enable_tpu" boolean NOT NULL DEFAULT false,
  "tpu_ipv4_cidr_block" character varying NULL,
  "addons_config_json" jsonb NULL,
  "private_cluster_config_json" jsonb NULL,
  "ip_allocation_policy_json" jsonb NULL,
  "network_config_json" jsonb NULL,
  "master_auth_json" jsonb NULL,
  "autoscaling_json" jsonb NULL,
  "vertical_pod_autoscaling_json" jsonb NULL,
  "monitoring_config_json" jsonb NULL,
  "logging_config_json" jsonb NULL,
  "maintenance_policy_json" jsonb NULL,
  "database_encryption_json" jsonb NULL,
  "workload_identity_config_json" jsonb NULL,
  "autopilot_json" jsonb NULL,
  "release_channel_json" jsonb NULL,
  "binary_authorization_json" jsonb NULL,
  "security_posture_config_json" jsonb NULL,
  "node_pool_defaults_json" jsonb NULL,
  "fleet_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcontainercluster_collected_at" to table: "gcp_container_clusters"
CREATE INDEX "bronzegcpcontainercluster_collected_at" ON "bronze"."gcp_container_clusters" ("collected_at");
-- Create index "bronzegcpcontainercluster_project_id" to table: "gcp_container_clusters"
CREATE INDEX "bronzegcpcontainercluster_project_id" ON "bronze"."gcp_container_clusters" ("project_id");
-- Create index "bronzegcpcontainercluster_status" to table: "gcp_container_clusters"
CREATE INDEX "bronzegcpcontainercluster_status" ON "bronze"."gcp_container_clusters" ("status");
-- Create "gcp_container_cluster_addons" table
CREATE TABLE "bronze"."gcp_container_cluster_addons" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "addon_name" character varying NOT NULL,
  "enabled" boolean NOT NULL DEFAULT false,
  "config_json" jsonb NULL,
  "bronze_gcp_container_cluster_addons" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_container_cluster_addons_gcp_container_clusters_addons" FOREIGN KEY ("bronze_gcp_container_cluster_addons") REFERENCES "bronze"."gcp_container_clusters" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_container_cluster_conditions" table
CREATE TABLE "bronze"."gcp_container_cluster_conditions" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "code" character varying NULL,
  "message" character varying NULL,
  "canonical_code" character varying NULL,
  "bronze_gcp_container_cluster_conditions" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_container_cluster_conditions_gcp_container_clusters_conditi" FOREIGN KEY ("bronze_gcp_container_cluster_conditions") REFERENCES "bronze"."gcp_container_clusters" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_container_cluster_labels" table
CREATE TABLE "bronze"."gcp_container_cluster_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_container_cluster_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_container_cluster_labels_gcp_container_clusters_labels" FOREIGN KEY ("bronze_gcp_container_cluster_labels") REFERENCES "bronze"."gcp_container_clusters" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_container_cluster_node_pools" table
CREATE TABLE "bronze"."gcp_container_cluster_node_pools" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "name" character varying NOT NULL,
  "version" character varying NULL,
  "status" character varying NULL,
  "status_message" character varying NULL,
  "initial_node_count" integer NULL,
  "self_link" character varying NULL,
  "pod_ipv4_cidr_size" integer NULL,
  "etag" character varying NULL,
  "locations_json" jsonb NULL,
  "config_json" jsonb NULL,
  "autoscaling_json" jsonb NULL,
  "management_json" jsonb NULL,
  "upgrade_settings_json" jsonb NULL,
  "network_config_json" jsonb NULL,
  "placement_policy_json" jsonb NULL,
  "max_pods_constraint_json" jsonb NULL,
  "bronze_gcp_container_cluster_node_pools" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_container_cluster_node_pools_gcp_container_clusters_node_po" FOREIGN KEY ("bronze_gcp_container_cluster_node_pools") REFERENCES "bronze"."gcp_container_clusters" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_projects" table
CREATE TABLE "bronze"."gcp_projects" (
  "project_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "project_number" character varying NOT NULL,
  "display_name" character varying NULL,
  "state" character varying NULL,
  "parent" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  "etag" character varying NULL,
  PRIMARY KEY ("project_id")
);
-- Create index "bronzegcpproject_collected_at" to table: "gcp_projects"
CREATE INDEX "bronzegcpproject_collected_at" ON "bronze"."gcp_projects" ("collected_at");
-- Create index "bronzegcpproject_parent" to table: "gcp_projects"
CREATE INDEX "bronzegcpproject_parent" ON "bronze"."gcp_projects" ("parent");
-- Create index "bronzegcpproject_state" to table: "gcp_projects"
CREATE INDEX "bronzegcpproject_state" ON "bronze"."gcp_projects" ("state");
-- Create index "gcp_projects_project_number_key" to table: "gcp_projects"
CREATE UNIQUE INDEX "gcp_projects_project_number_key" ON "bronze"."gcp_projects" ("project_number");
-- Create "gcp_project_labels" table
CREATE TABLE "bronze"."gcp_project_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_project_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_project_labels_gcp_projects_labels" FOREIGN KEY ("bronze_gcp_project_labels") REFERENCES "bronze"."gcp_projects" ("project_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
