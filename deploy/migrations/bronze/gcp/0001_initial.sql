-- Create "gcp_logging_log_metrics" table
CREATE TABLE "bronze"."gcp_logging_log_metrics" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "filter" text NULL,
  "metric_descriptor_json" jsonb NULL,
  "label_extractors_json" jsonb NULL,
  "bucket_options_json" jsonb NULL,
  "value_extractor" character varying NULL,
  "version" character varying NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcplogginglogmetric_collected_at" to table: "gcp_logging_log_metrics"
CREATE INDEX "bronzegcplogginglogmetric_collected_at" ON "bronze"."gcp_logging_log_metrics" ("collected_at");
-- Create index "bronzegcplogginglogmetric_project_id" to table: "gcp_logging_log_metrics"
CREATE INDEX "bronzegcplogginglogmetric_project_id" ON "bronze"."gcp_logging_log_metrics" ("project_id");
-- Create "gcp_monitoring_alert_policies" table
CREATE TABLE "bronze"."gcp_monitoring_alert_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "documentation_json" jsonb NULL,
  "user_labels_json" jsonb NULL,
  "conditions_json" jsonb NULL,
  "combiner" bigint NOT NULL DEFAULT 0,
  "enabled" boolean NOT NULL DEFAULT false,
  "notification_channels_json" jsonb NULL,
  "creation_record_json" jsonb NULL,
  "mutation_record_json" jsonb NULL,
  "alert_strategy_json" jsonb NULL,
  "severity" bigint NOT NULL DEFAULT 0,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpmonitoringalertpolicy_collected_at" to table: "gcp_monitoring_alert_policies"
CREATE INDEX "bronzegcpmonitoringalertpolicy_collected_at" ON "bronze"."gcp_monitoring_alert_policies" ("collected_at");
-- Create index "bronzegcpmonitoringalertpolicy_project_id" to table: "gcp_monitoring_alert_policies"
CREATE INDEX "bronzegcpmonitoringalertpolicy_project_id" ON "bronze"."gcp_monitoring_alert_policies" ("project_id");
-- Create "gcp_alloydb_clusters" table
CREATE TABLE "bronze"."gcp_alloydb_clusters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "uid" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  "labels_json" jsonb NULL,
  "state" bigint NULL,
  "cluster_type" bigint NULL,
  "database_version" bigint NULL,
  "network_config_json" jsonb NULL,
  "network" character varying NULL,
  "etag" character varying NULL,
  "annotations_json" jsonb NULL,
  "reconciling" boolean NULL,
  "initial_user_json" jsonb NULL,
  "automated_backup_policy_json" jsonb NULL,
  "ssl_config_json" jsonb NULL,
  "encryption_config_json" jsonb NULL,
  "encryption_info_json" jsonb NULL,
  "continuous_backup_config_json" jsonb NULL,
  "continuous_backup_info_json" jsonb NULL,
  "secondary_config_json" jsonb NULL,
  "primary_config_json" jsonb NULL,
  "satisfies_pzs" boolean NULL,
  "psc_config_json" jsonb NULL,
  "maintenance_update_policy_json" jsonb NULL,
  "maintenance_schedule_json" jsonb NULL,
  "subscription_type" bigint NULL,
  "trial_metadata_json" jsonb NULL,
  "tags_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpalloydbcluster_collected_at" to table: "gcp_alloydb_clusters"
CREATE INDEX "bronzegcpalloydbcluster_collected_at" ON "bronze"."gcp_alloydb_clusters" ("collected_at");
-- Create index "bronzegcpalloydbcluster_project_id" to table: "gcp_alloydb_clusters"
CREATE INDEX "bronzegcpalloydbcluster_project_id" ON "bronze"."gcp_alloydb_clusters" ("project_id");
-- Create "gcp_bigquery_datasets" table
CREATE TABLE "bronze"."gcp_bigquery_datasets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "friendly_name" character varying NULL,
  "description" character varying NULL,
  "location" character varying NULL,
  "default_table_expiration_ms" bigint NULL,
  "default_partition_expiration_ms" bigint NULL,
  "labels_json" jsonb NULL,
  "access_json" jsonb NULL,
  "creation_time" character varying NULL,
  "last_modified_time" character varying NULL,
  "etag" character varying NULL,
  "default_collation" character varying NULL,
  "max_time_travel_hours" bigint NULL,
  "default_encryption_configuration_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpbigquerydataset_collected_at" to table: "gcp_bigquery_datasets"
CREATE INDEX "bronzegcpbigquerydataset_collected_at" ON "bronze"."gcp_bigquery_datasets" ("collected_at");
-- Create index "bronzegcpbigquerydataset_project_id" to table: "gcp_bigquery_datasets"
CREATE INDEX "bronzegcpbigquerydataset_project_id" ON "bronze"."gcp_bigquery_datasets" ("project_id");
-- Create "gcp_bigquery_tables" table
CREATE TABLE "bronze"."gcp_bigquery_tables" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "dataset_id" character varying NOT NULL,
  "friendly_name" character varying NULL,
  "description" character varying NULL,
  "schema_json" jsonb NULL,
  "num_bytes" bigint NULL,
  "num_long_term_bytes" bigint NULL,
  "num_rows" bigint NULL,
  "creation_time" character varying NULL,
  "expiration_time" character varying NULL,
  "last_modified_time" character varying NULL,
  "table_type" character varying NULL,
  "labels_json" jsonb NULL,
  "encryption_configuration_json" jsonb NULL,
  "time_partitioning_json" jsonb NULL,
  "range_partitioning_json" jsonb NULL,
  "clustering_json" jsonb NULL,
  "require_partition_filter" boolean NOT NULL DEFAULT false,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpbigquerytable_collected_at" to table: "gcp_bigquery_tables"
CREATE INDEX "bronzegcpbigquerytable_collected_at" ON "bronze"."gcp_bigquery_tables" ("collected_at");
-- Create index "bronzegcpbigquerytable_dataset_id" to table: "gcp_bigquery_tables"
CREATE INDEX "bronzegcpbigquerytable_dataset_id" ON "bronze"."gcp_bigquery_tables" ("dataset_id");
-- Create index "bronzegcpbigquerytable_project_id" to table: "gcp_bigquery_tables"
CREATE INDEX "bronzegcpbigquerytable_project_id" ON "bronze"."gcp_bigquery_tables" ("project_id");
-- Create "gcp_binaryauthorization_attestors" table
CREATE TABLE "bronze"."gcp_binaryauthorization_attestors" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "description" character varying NULL,
  "user_owned_grafeas_note_json" jsonb NULL,
  "update_time" character varying NULL,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpbinaryauthorizationattestor_collected_at" to table: "gcp_binaryauthorization_attestors"
CREATE INDEX "bronzegcpbinaryauthorizationattestor_collected_at" ON "bronze"."gcp_binaryauthorization_attestors" ("collected_at");
-- Create index "bronzegcpbinaryauthorizationattestor_project_id" to table: "gcp_binaryauthorization_attestors"
CREATE INDEX "bronzegcpbinaryauthorizationattestor_project_id" ON "bronze"."gcp_binaryauthorization_attestors" ("project_id");
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
-- Create "gcp_compute_security_policies" table
CREATE TABLE "bronze"."gcp_compute_security_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "type" character varying NULL,
  "fingerprint" character varying NULL,
  "rules_json" jsonb NULL,
  "associations_json" jsonb NULL,
  "adaptive_protection_config_json" jsonb NULL,
  "advanced_options_config_json" jsonb NULL,
  "ddos_protection_config_json" jsonb NULL,
  "recaptcha_options_config_json" jsonb NULL,
  "labels_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputesecuritypolicy_collected_at" to table: "gcp_compute_security_policies"
CREATE INDEX "bronzegcpcomputesecuritypolicy_collected_at" ON "bronze"."gcp_compute_security_policies" ("collected_at");
-- Create index "bronzegcpcomputesecuritypolicy_project_id" to table: "gcp_compute_security_policies"
CREATE INDEX "bronzegcpcomputesecuritypolicy_project_id" ON "bronze"."gcp_compute_security_policies" ("project_id");
-- Create "gcp_binaryauthorization_policies" table
CREATE TABLE "bronze"."gcp_binaryauthorization_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "description" character varying NULL,
  "global_policy_evaluation_mode" bigint NOT NULL DEFAULT 0,
  "default_admission_rule_json" jsonb NULL,
  "cluster_admission_rules_json" jsonb NULL,
  "kube_namespace_admission_rules_json" jsonb NULL,
  "istio_service_identity_admission_rules_json" jsonb NULL,
  "update_time" character varying NULL,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpbinaryauthorizationpolicy_collected_at" to table: "gcp_binaryauthorization_policies"
CREATE INDEX "bronzegcpbinaryauthorizationpolicy_collected_at" ON "bronze"."gcp_binaryauthorization_policies" ("collected_at");
-- Create index "bronzegcpbinaryauthorizationpolicy_project_id" to table: "gcp_binaryauthorization_policies"
CREATE INDEX "bronzegcpbinaryauthorizationpolicy_project_id" ON "bronze"."gcp_binaryauthorization_policies" ("project_id");
-- Create "gcp_cloudasset_assets" table
CREATE TABLE "bronze"."gcp_cloudasset_assets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "asset_type" character varying NOT NULL,
  "organization_id" character varying NOT NULL,
  "update_time" character varying NULL,
  "resource_json" jsonb NULL,
  "iam_policy_json" jsonb NULL,
  "org_policy_json" jsonb NULL,
  "access_policy_json" jsonb NULL,
  "os_inventory_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcloudassetasset_asset_type" to table: "gcp_cloudasset_assets"
CREATE INDEX "bronzegcpcloudassetasset_asset_type" ON "bronze"."gcp_cloudasset_assets" ("asset_type");
-- Create index "bronzegcpcloudassetasset_collected_at" to table: "gcp_cloudasset_assets"
CREATE INDEX "bronzegcpcloudassetasset_collected_at" ON "bronze"."gcp_cloudasset_assets" ("collected_at");
-- Create index "bronzegcpcloudassetasset_organization_id" to table: "gcp_cloudasset_assets"
CREATE INDEX "bronzegcpcloudassetasset_organization_id" ON "bronze"."gcp_cloudasset_assets" ("organization_id");
-- Create "gcp_cloudasset_iam_policy_searches" table
CREATE TABLE "bronze"."gcp_cloudasset_iam_policy_searches" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "asset_type" character varying NULL,
  "project" character varying NULL,
  "organization" character varying NULL,
  "organization_id" character varying NOT NULL,
  "folders_json" jsonb NULL,
  "policy_json" jsonb NULL,
  "explanation_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcloudassetiampolicysearch_asset_type" to table: "gcp_cloudasset_iam_policy_searches"
CREATE INDEX "bronzegcpcloudassetiampolicysearch_asset_type" ON "bronze"."gcp_cloudasset_iam_policy_searches" ("asset_type");
-- Create index "bronzegcpcloudassetiampolicysearch_collected_at" to table: "gcp_cloudasset_iam_policy_searches"
CREATE INDEX "bronzegcpcloudassetiampolicysearch_collected_at" ON "bronze"."gcp_cloudasset_iam_policy_searches" ("collected_at");
-- Create index "bronzegcpcloudassetiampolicysearch_organization_id" to table: "gcp_cloudasset_iam_policy_searches"
CREATE INDEX "bronzegcpcloudassetiampolicysearch_organization_id" ON "bronze"."gcp_cloudasset_iam_policy_searches" ("organization_id");
-- Create "gcp_cloudasset_resource_searches" table
CREATE TABLE "bronze"."gcp_cloudasset_resource_searches" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "asset_type" character varying NOT NULL,
  "project" character varying NULL,
  "display_name" character varying NULL,
  "description" character varying NULL,
  "location" character varying NULL,
  "organization_id" character varying NOT NULL,
  "labels_json" jsonb NULL,
  "network_tags_json" jsonb NULL,
  "additional_attributes_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcloudassetresourcesearch_asset_type" to table: "gcp_cloudasset_resource_searches"
CREATE INDEX "bronzegcpcloudassetresourcesearch_asset_type" ON "bronze"."gcp_cloudasset_resource_searches" ("asset_type");
-- Create index "bronzegcpcloudassetresourcesearch_collected_at" to table: "gcp_cloudasset_resource_searches"
CREATE INDEX "bronzegcpcloudassetresourcesearch_collected_at" ON "bronze"."gcp_cloudasset_resource_searches" ("collected_at");
-- Create index "bronzegcpcloudassetresourcesearch_location" to table: "gcp_cloudasset_resource_searches"
CREATE INDEX "bronzegcpcloudassetresourcesearch_location" ON "bronze"."gcp_cloudasset_resource_searches" ("location");
-- Create index "bronzegcpcloudassetresourcesearch_organization_id" to table: "gcp_cloudasset_resource_searches"
CREATE INDEX "bronzegcpcloudassetresourcesearch_organization_id" ON "bronze"."gcp_cloudasset_resource_searches" ("organization_id");
-- Create index "bronzegcpcloudassetresourcesearch_project" to table: "gcp_cloudasset_resource_searches"
CREATE INDEX "bronzegcpcloudassetresourcesearch_project" ON "bronze"."gcp_cloudasset_resource_searches" ("project");
-- Create "gcp_cloudfunctions_functions" table
CREATE TABLE "bronze"."gcp_cloudfunctions_functions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "environment" bigint NOT NULL DEFAULT 0,
  "state" bigint NOT NULL DEFAULT 0,
  "build_config_json" jsonb NULL,
  "service_config_json" jsonb NULL,
  "event_trigger_json" jsonb NULL,
  "state_messages_json" jsonb NULL,
  "update_time" character varying NULL,
  "create_time" character varying NULL,
  "labels_json" jsonb NULL,
  "kms_key_name" character varying NULL,
  "url" character varying NULL,
  "satisfies_pzs" boolean NOT NULL DEFAULT false,
  "project_id" character varying NOT NULL,
  "location" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcloudfunctionsfunction_collected_at" to table: "gcp_cloudfunctions_functions"
CREATE INDEX "bronzegcpcloudfunctionsfunction_collected_at" ON "bronze"."gcp_cloudfunctions_functions" ("collected_at");
-- Create index "bronzegcpcloudfunctionsfunction_project_id" to table: "gcp_cloudfunctions_functions"
CREATE INDEX "bronzegcpcloudfunctionsfunction_project_id" ON "bronze"."gcp_cloudfunctions_functions" ("project_id");
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
-- Create "gcp_compute_routers" table
CREATE TABLE "bronze"."gcp_compute_routers" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "network" character varying NULL,
  "region" character varying NULL,
  "bgp_asn" bigint NOT NULL DEFAULT 0,
  "bgp_advertise_mode" character varying NULL,
  "bgp_advertised_groups_json" jsonb NULL,
  "bgp_advertised_ip_ranges_json" jsonb NULL,
  "bgp_keepalive_interval" bigint NOT NULL DEFAULT 0,
  "bgp_peers_json" jsonb NULL,
  "interfaces_json" jsonb NULL,
  "nats_json" jsonb NULL,
  "encrypted_interconnect_router" boolean NOT NULL DEFAULT false,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputerouter_collected_at" to table: "gcp_compute_routers"
CREATE INDEX "bronzegcpcomputerouter_collected_at" ON "bronze"."gcp_compute_routers" ("collected_at");
-- Create index "bronzegcpcomputerouter_project_id" to table: "gcp_compute_routers"
CREATE INDEX "bronzegcpcomputerouter_project_id" ON "bronze"."gcp_compute_routers" ("project_id");
-- Create "gcp_accesscontextmanager_access_policies" table
CREATE TABLE "bronze"."gcp_accesscontextmanager_access_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "parent" character varying NOT NULL,
  "title" character varying NULL,
  "etag" character varying NULL,
  "scopes_json" jsonb NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpaccesscontextmanageraccesspolicy_collected_at" to table: "gcp_accesscontextmanager_access_policies"
CREATE INDEX "bronzegcpaccesscontextmanageraccesspolicy_collected_at" ON "bronze"."gcp_accesscontextmanager_access_policies" ("collected_at");
-- Create index "bronzegcpaccesscontextmanageraccesspolicy_organization_id" to table: "gcp_accesscontextmanager_access_policies"
CREATE INDEX "bronzegcpaccesscontextmanageraccesspolicy_organization_id" ON "bronze"."gcp_accesscontextmanager_access_policies" ("organization_id");
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
-- Create "gcp_serviceusage_enabled_services" table
CREATE TABLE "bronze"."gcp_serviceusage_enabled_services" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "parent" character varying NOT NULL,
  "config_json" jsonb NULL,
  "state" bigint NOT NULL DEFAULT 0,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpserviceusageenabledservice_collected_at" to table: "gcp_serviceusage_enabled_services"
CREATE INDEX "bronzegcpserviceusageenabledservice_collected_at" ON "bronze"."gcp_serviceusage_enabled_services" ("collected_at");
-- Create index "bronzegcpserviceusageenabledservice_project_id" to table: "gcp_serviceusage_enabled_services"
CREATE INDEX "bronzegcpserviceusageenabledservice_project_id" ON "bronze"."gcp_serviceusage_enabled_services" ("project_id");
-- Create "gcp_securitycenter_sources" table
CREATE TABLE "bronze"."gcp_securitycenter_sources" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "display_name" character varying NULL,
  "description" character varying NULL,
  "canonical_name" character varying NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpsecuritycentersource_collected_at" to table: "gcp_securitycenter_sources"
CREATE INDEX "bronzegcpsecuritycentersource_collected_at" ON "bronze"."gcp_securitycenter_sources" ("collected_at");
-- Create index "bronzegcpsecuritycentersource_organization_id" to table: "gcp_securitycenter_sources"
CREATE INDEX "bronzegcpsecuritycentersource_organization_id" ON "bronze"."gcp_securitycenter_sources" ("organization_id");
-- Create "gcp_securitycenter_notification_configs" table
CREATE TABLE "bronze"."gcp_securitycenter_notification_configs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "pubsub_topic" character varying NULL,
  "streaming_config_json" character varying NULL,
  "service_account" character varying NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpsecuritycenternotificationconfig_collected_at" to table: "gcp_securitycenter_notification_configs"
CREATE INDEX "bronzegcpsecuritycenternotificationconfig_collected_at" ON "bronze"."gcp_securitycenter_notification_configs" ("collected_at");
-- Create index "bronzegcpsecuritycenternotificationconfig_organization_id" to table: "gcp_securitycenter_notification_configs"
CREATE INDEX "bronzegcpsecuritycenternotificationconfig_organization_id" ON "bronze"."gcp_securitycenter_notification_configs" ("organization_id");
-- Create "gcp_securitycenter_findings" table
CREATE TABLE "bronze"."gcp_securitycenter_findings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "parent" character varying NOT NULL,
  "resource_name" character varying NULL,
  "state" character varying NULL,
  "category" character varying NULL,
  "external_uri" character varying NULL,
  "severity" character varying NULL,
  "finding_class" character varying NULL,
  "canonical_name" character varying NULL,
  "mute" character varying NULL,
  "organization_id" character varying NOT NULL,
  "create_time" character varying NULL,
  "event_time" character varying NULL,
  "source_properties" jsonb NULL,
  "security_marks" jsonb NULL,
  "indicator" jsonb NULL,
  "vulnerability" jsonb NULL,
  "connections" jsonb NULL,
  "compliances" jsonb NULL,
  "contacts" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpsecuritycenterfinding_category" to table: "gcp_securitycenter_findings"
CREATE INDEX "bronzegcpsecuritycenterfinding_category" ON "bronze"."gcp_securitycenter_findings" ("category");
-- Create index "bronzegcpsecuritycenterfinding_collected_at" to table: "gcp_securitycenter_findings"
CREATE INDEX "bronzegcpsecuritycenterfinding_collected_at" ON "bronze"."gcp_securitycenter_findings" ("collected_at");
-- Create index "bronzegcpsecuritycenterfinding_organization_id" to table: "gcp_securitycenter_findings"
CREATE INDEX "bronzegcpsecuritycenterfinding_organization_id" ON "bronze"."gcp_securitycenter_findings" ("organization_id");
-- Create index "bronzegcpsecuritycenterfinding_severity" to table: "gcp_securitycenter_findings"
CREATE INDEX "bronzegcpsecuritycenterfinding_severity" ON "bronze"."gcp_securitycenter_findings" ("severity");
-- Create index "bronzegcpsecuritycenterfinding_state" to table: "gcp_securitycenter_findings"
CREATE INDEX "bronzegcpsecuritycenterfinding_state" ON "bronze"."gcp_securitycenter_findings" ("state");
-- Create "gcp_redis_instances" table
CREATE TABLE "bronze"."gcp_redis_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "labels_json" jsonb NULL,
  "location_id" character varying NULL,
  "alternative_location_id" character varying NULL,
  "redis_version" character varying NULL,
  "reserved_ip_range" character varying NULL,
  "secondary_ip_range" character varying NULL,
  "host" character varying NULL,
  "port" integer NULL,
  "current_location_id" character varying NULL,
  "create_time" character varying NULL,
  "state" integer NULL,
  "status_message" character varying NULL,
  "redis_configs_json" jsonb NULL,
  "tier" integer NULL,
  "memory_size_gb" integer NULL,
  "authorized_network" character varying NULL,
  "persistence_iam_identity" character varying NULL,
  "connect_mode" integer NULL,
  "auth_enabled" boolean NOT NULL DEFAULT false,
  "server_ca_certs_json" jsonb NULL,
  "transit_encryption_mode" integer NULL,
  "maintenance_policy_json" jsonb NULL,
  "maintenance_schedule_json" jsonb NULL,
  "replica_count" integer NULL,
  "nodes_json" jsonb NULL,
  "read_endpoint" character varying NULL,
  "read_endpoint_port" integer NULL,
  "read_replicas_mode" integer NULL,
  "customer_managed_key" character varying NULL,
  "persistence_config_json" jsonb NULL,
  "suspension_reasons_json" jsonb NULL,
  "maintenance_version" character varying NULL,
  "available_maintenance_versions_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpredisinstance_collected_at" to table: "gcp_redis_instances"
CREATE INDEX "bronzegcpredisinstance_collected_at" ON "bronze"."gcp_redis_instances" ("collected_at");
-- Create index "bronzegcpredisinstance_project_id" to table: "gcp_redis_instances"
CREATE INDEX "bronzegcpredisinstance_project_id" ON "bronze"."gcp_redis_instances" ("project_id");
-- Create "gcp_pubsub_topics" table
CREATE TABLE "bronze"."gcp_pubsub_topics" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "labels_json" jsonb NULL,
  "message_storage_policy_json" jsonb NULL,
  "kms_key_name" character varying NULL,
  "schema_settings_json" jsonb NULL,
  "message_retention_duration" character varying NULL,
  "state" bigint NULL,
  "ingestion_data_source_settings_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcppubsubtopic_collected_at" to table: "gcp_pubsub_topics"
CREATE INDEX "bronzegcppubsubtopic_collected_at" ON "bronze"."gcp_pubsub_topics" ("collected_at");
-- Create index "bronzegcppubsubtopic_project_id" to table: "gcp_pubsub_topics"
CREATE INDEX "bronzegcppubsubtopic_project_id" ON "bronze"."gcp_pubsub_topics" ("project_id");
-- Create "gcp_pubsub_subscriptions" table
CREATE TABLE "bronze"."gcp_pubsub_subscriptions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "topic" character varying NULL,
  "push_config_json" jsonb NULL,
  "bigquery_config_json" jsonb NULL,
  "cloud_storage_config_json" jsonb NULL,
  "ack_deadline_seconds" bigint NULL,
  "retain_acked_messages" boolean NOT NULL DEFAULT false,
  "message_retention_duration" character varying NULL,
  "labels_json" jsonb NULL,
  "enable_message_ordering" boolean NOT NULL DEFAULT false,
  "expiration_policy_json" jsonb NULL,
  "filter" character varying NULL,
  "dead_letter_policy_json" jsonb NULL,
  "retry_policy_json" jsonb NULL,
  "detached" boolean NOT NULL DEFAULT false,
  "enable_exactly_once_delivery" boolean NOT NULL DEFAULT false,
  "state" bigint NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcppubsubsubscription_collected_at" to table: "gcp_pubsub_subscriptions"
CREATE INDEX "bronzegcppubsubsubscription_collected_at" ON "bronze"."gcp_pubsub_subscriptions" ("collected_at");
-- Create index "bronzegcppubsubsubscription_project_id" to table: "gcp_pubsub_subscriptions"
CREATE INDEX "bronzegcppubsubsubscription_project_id" ON "bronze"."gcp_pubsub_subscriptions" ("project_id");
-- Create "gcp_orgpolicy_policies" table
CREATE TABLE "bronze"."gcp_orgpolicy_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "etag" character varying NULL,
  "spec" jsonb NULL,
  "dry_run_spec" jsonb NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcporgpolicypolicy_collected_at" to table: "gcp_orgpolicy_policies"
CREATE INDEX "bronzegcporgpolicypolicy_collected_at" ON "bronze"."gcp_orgpolicy_policies" ("collected_at");
-- Create index "bronzegcporgpolicypolicy_organization_id" to table: "gcp_orgpolicy_policies"
CREATE INDEX "bronzegcporgpolicypolicy_organization_id" ON "bronze"."gcp_orgpolicy_policies" ("organization_id");
-- Create "gcp_orgpolicy_custom_constraints" table
CREATE TABLE "bronze"."gcp_orgpolicy_custom_constraints" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_types" jsonb NULL,
  "method_types" jsonb NULL,
  "condition" character varying NULL,
  "action_type" bigint NOT NULL DEFAULT 0,
  "display_name" character varying NULL,
  "description" character varying NULL,
  "update_time" timestamptz NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcporgpolicycustomconstraint_collected_at" to table: "gcp_orgpolicy_custom_constraints"
CREATE INDEX "bronzegcporgpolicycustomconstraint_collected_at" ON "bronze"."gcp_orgpolicy_custom_constraints" ("collected_at");
-- Create index "bronzegcporgpolicycustomconstraint_organization_id" to table: "gcp_orgpolicy_custom_constraints"
CREATE INDEX "bronzegcporgpolicycustomconstraint_organization_id" ON "bronze"."gcp_orgpolicy_custom_constraints" ("organization_id");
-- Create "gcp_orgpolicy_constraints" table
CREATE TABLE "bronze"."gcp_orgpolicy_constraints" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "display_name" character varying NULL,
  "description" character varying NULL,
  "constraint_default" bigint NOT NULL DEFAULT 0,
  "supports_dry_run" boolean NOT NULL DEFAULT false,
  "supports_simulation" boolean NOT NULL DEFAULT false,
  "list_constraint" jsonb NULL,
  "boolean_constraint" jsonb NULL,
  "organization_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcporgpolicyconstraint_collected_at" to table: "gcp_orgpolicy_constraints"
CREATE INDEX "bronzegcporgpolicyconstraint_collected_at" ON "bronze"."gcp_orgpolicy_constraints" ("collected_at");
-- Create index "bronzegcporgpolicyconstraint_organization_id" to table: "gcp_orgpolicy_constraints"
CREATE INDEX "bronzegcporgpolicyconstraint_organization_id" ON "bronze"."gcp_orgpolicy_constraints" ("organization_id");
-- Create "gcp_organizations" table
CREATE TABLE "bronze"."gcp_organizations" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "state" character varying NULL,
  "directory_customer_id" character varying NULL,
  "etag" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcporganization_collected_at" to table: "gcp_organizations"
CREATE INDEX "bronzegcporganization_collected_at" ON "bronze"."gcp_organizations" ("collected_at");
-- Create index "bronzegcporganization_state" to table: "gcp_organizations"
CREATE INDEX "bronzegcporganization_state" ON "bronze"."gcp_organizations" ("state");
-- Create "gcp_monitoring_uptime_check_configs" table
CREATE TABLE "bronze"."gcp_monitoring_uptime_check_configs" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "monitored_resource_json" jsonb NULL,
  "resource_group_json" jsonb NULL,
  "http_check_json" jsonb NULL,
  "tcp_check_json" jsonb NULL,
  "period" character varying NULL,
  "timeout" character varying NULL,
  "content_matchers_json" jsonb NULL,
  "checker_type" bigint NOT NULL DEFAULT 0,
  "selected_regions_json" jsonb NULL,
  "is_internal" boolean NOT NULL DEFAULT false,
  "internal_checkers_json" jsonb NULL,
  "user_labels_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpmonitoringuptimecheckconfig_collected_at" to table: "gcp_monitoring_uptime_check_configs"
CREATE INDEX "bronzegcpmonitoringuptimecheckconfig_collected_at" ON "bronze"."gcp_monitoring_uptime_check_configs" ("collected_at");
-- Create index "bronzegcpmonitoringuptimecheckconfig_project_id" to table: "gcp_monitoring_uptime_check_configs"
CREATE INDEX "bronzegcpmonitoringuptimecheckconfig_project_id" ON "bronze"."gcp_monitoring_uptime_check_configs" ("project_id");
-- Create "gcp_compute_ssl_policies" table
CREATE TABLE "bronze"."gcp_compute_ssl_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "profile" character varying NULL,
  "min_tls_version" character varying NULL,
  "fingerprint" character varying NULL,
  "custom_features_json" jsonb NULL,
  "enabled_features_json" jsonb NULL,
  "warnings_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputesslpolicy_collected_at" to table: "gcp_compute_ssl_policies"
CREATE INDEX "bronzegcpcomputesslpolicy_collected_at" ON "bronze"."gcp_compute_ssl_policies" ("collected_at");
-- Create index "bronzegcpcomputesslpolicy_project_id" to table: "gcp_compute_ssl_policies"
CREATE INDEX "bronzegcpcomputesslpolicy_project_id" ON "bronze"."gcp_compute_ssl_policies" ("project_id");
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
-- Create "gcp_logging_sinks" table
CREATE TABLE "bronze"."gcp_logging_sinks" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "destination" character varying NULL,
  "filter" text NULL,
  "description" character varying NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "include_children" boolean NOT NULL DEFAULT false,
  "writer_identity" character varying NULL,
  "exclusions_json" jsonb NULL,
  "bigquery_options_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcploggingsink_collected_at" to table: "gcp_logging_sinks"
CREATE INDEX "bronzegcploggingsink_collected_at" ON "bronze"."gcp_logging_sinks" ("collected_at");
-- Create index "bronzegcploggingsink_project_id" to table: "gcp_logging_sinks"
CREATE INDEX "bronzegcploggingsink_project_id" ON "bronze"."gcp_logging_sinks" ("project_id");
-- Create "gcp_compute_packet_mirrorings" table
CREATE TABLE "bronze"."gcp_compute_packet_mirrorings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "region" character varying NULL,
  "network" character varying NULL,
  "priority" bigint NULL,
  "enable" character varying NULL,
  "collector_ilb_json" jsonb NULL,
  "mirrored_resources_json" jsonb NULL,
  "filter_json" jsonb NULL,
  "creation_timestamp" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputepacketmirroring_collected_at" to table: "gcp_compute_packet_mirrorings"
CREATE INDEX "bronzegcpcomputepacketmirroring_collected_at" ON "bronze"."gcp_compute_packet_mirrorings" ("collected_at");
-- Create index "bronzegcpcomputepacketmirroring_project_id" to table: "gcp_compute_packet_mirrorings"
CREATE INDEX "bronzegcpcomputepacketmirroring_project_id" ON "bronze"."gcp_compute_packet_mirrorings" ("project_id");
-- Create "gcp_logging_log_exclusions" table
CREATE TABLE "bronze"."gcp_logging_log_exclusions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "filter" text NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcplogginglogexclusion_collected_at" to table: "gcp_logging_log_exclusions"
CREATE INDEX "bronzegcplogginglogexclusion_collected_at" ON "bronze"."gcp_logging_log_exclusions" ("collected_at");
-- Create index "bronzegcplogginglogexclusion_project_id" to table: "gcp_logging_log_exclusions"
CREATE INDEX "bronzegcplogginglogexclusion_project_id" ON "bronze"."gcp_logging_log_exclusions" ("project_id");
-- Create "gcp_logging_buckets" table
CREATE TABLE "bronze"."gcp_logging_buckets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "retention_days" integer NOT NULL DEFAULT 0,
  "locked" boolean NOT NULL DEFAULT false,
  "lifecycle_state" character varying NULL,
  "analytics_enabled" boolean NOT NULL DEFAULT false,
  "project_id" character varying NOT NULL,
  "location" character varying NULL,
  "cmek_settings_json" jsonb NULL,
  "index_configs_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcploggingbucket_collected_at" to table: "gcp_logging_buckets"
CREATE INDEX "bronzegcploggingbucket_collected_at" ON "bronze"."gcp_logging_buckets" ("collected_at");
-- Create index "bronzegcploggingbucket_project_id" to table: "gcp_logging_buckets"
CREATE INDEX "bronzegcploggingbucket_project_id" ON "bronze"."gcp_logging_buckets" ("project_id");
-- Create "gcp_kms_key_rings" table
CREATE TABLE "bronze"."gcp_kms_key_rings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "create_time" character varying NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpkmskeyring_collected_at" to table: "gcp_kms_key_rings"
CREATE INDEX "bronzegcpkmskeyring_collected_at" ON "bronze"."gcp_kms_key_rings" ("collected_at");
-- Create index "bronzegcpkmskeyring_project_id" to table: "gcp_kms_key_rings"
CREATE INDEX "bronzegcpkmskeyring_project_id" ON "bronze"."gcp_kms_key_rings" ("project_id");
-- Create "gcp_kms_crypto_keys" table
CREATE TABLE "bronze"."gcp_kms_crypto_keys" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "purpose" character varying NULL,
  "create_time" character varying NULL,
  "next_rotation_time" character varying NULL,
  "rotation_period" character varying NULL,
  "destroy_scheduled_duration" character varying NULL,
  "import_only" boolean NOT NULL DEFAULT false,
  "crypto_key_backend" character varying NULL,
  "version_template_json" jsonb NULL,
  "primary_json" jsonb NULL,
  "labels_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NULL,
  "key_ring_name" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpkmscryptokey_collected_at" to table: "gcp_kms_crypto_keys"
CREATE INDEX "bronzegcpkmscryptokey_collected_at" ON "bronze"."gcp_kms_crypto_keys" ("collected_at");
-- Create index "bronzegcpkmscryptokey_project_id" to table: "gcp_kms_crypto_keys"
CREATE INDEX "bronzegcpkmscryptokey_project_id" ON "bronze"."gcp_kms_crypto_keys" ("project_id");
-- Create "gcp_iap_settings" table
CREATE TABLE "bronze"."gcp_iap_settings" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "access_settings_json" jsonb NULL,
  "application_settings_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpiapsettings_collected_at" to table: "gcp_iap_settings"
CREATE INDEX "bronzegcpiapsettings_collected_at" ON "bronze"."gcp_iap_settings" ("collected_at");
-- Create index "bronzegcpiapsettings_project_id" to table: "gcp_iap_settings"
CREATE INDEX "bronzegcpiapsettings_project_id" ON "bronze"."gcp_iap_settings" ("project_id");
-- Create "gcp_iap_iam_policies" table
CREATE TABLE "bronze"."gcp_iap_iam_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "etag" character varying NULL,
  "version" bigint NULL,
  "bindings_json" jsonb NULL,
  "audit_configs_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpiapiampolicy_collected_at" to table: "gcp_iap_iam_policies"
CREATE INDEX "bronzegcpiapiampolicy_collected_at" ON "bronze"."gcp_iap_iam_policies" ("collected_at");
-- Create index "bronzegcpiapiampolicy_project_id" to table: "gcp_iap_iam_policies"
CREATE INDEX "bronzegcpiapiampolicy_project_id" ON "bronze"."gcp_iap_iam_policies" ("project_id");
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
-- Create "gcp_filestore_instances" table
CREATE TABLE "bronze"."gcp_filestore_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "state" bigint NOT NULL DEFAULT 0,
  "status_message" character varying NULL,
  "create_time" character varying NULL,
  "tier" bigint NOT NULL DEFAULT 0,
  "labels_json" jsonb NULL,
  "file_shares_json" jsonb NULL,
  "networks_json" jsonb NULL,
  "etag" character varying NULL,
  "satisfies_pzs" boolean NOT NULL DEFAULT false,
  "satisfies_pzi" boolean NOT NULL DEFAULT false,
  "kms_key_name" character varying NULL,
  "suspension_reasons_json" jsonb NULL,
  "max_capacity_gb" bigint NULL,
  "protocol" bigint NOT NULL DEFAULT 0,
  "project_id" character varying NOT NULL,
  "location" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpfilestoreinstance_collected_at" to table: "gcp_filestore_instances"
CREATE INDEX "bronzegcpfilestoreinstance_collected_at" ON "bronze"."gcp_filestore_instances" ("collected_at");
-- Create index "bronzegcpfilestoreinstance_project_id" to table: "gcp_filestore_instances"
CREATE INDEX "bronzegcpfilestoreinstance_project_id" ON "bronze"."gcp_filestore_instances" ("project_id");
-- Create "gcp_dns_policies" table
CREATE TABLE "bronze"."gcp_dns_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "enable_inbound_forwarding" boolean NOT NULL DEFAULT false,
  "enable_logging" boolean NOT NULL DEFAULT false,
  "networks_json" jsonb NULL,
  "alternative_name_server_config_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpdnspolicy_collected_at" to table: "gcp_dns_policies"
CREATE INDEX "bronzegcpdnspolicy_collected_at" ON "bronze"."gcp_dns_policies" ("collected_at");
-- Create index "bronzegcpdnspolicy_project_id" to table: "gcp_dns_policies"
CREATE INDEX "bronzegcpdnspolicy_project_id" ON "bronze"."gcp_dns_policies" ("project_id");
-- Create "gcp_dataproc_clusters" table
CREATE TABLE "bronze"."gcp_dataproc_clusters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "cluster_name" character varying NOT NULL,
  "cluster_uuid" character varying NULL,
  "config_json" jsonb NULL,
  "status_json" jsonb NULL,
  "status_history_json" jsonb NULL,
  "labels_json" jsonb NULL,
  "metrics_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpdataproccluster_collected_at" to table: "gcp_dataproc_clusters"
CREATE INDEX "bronzegcpdataproccluster_collected_at" ON "bronze"."gcp_dataproc_clusters" ("collected_at");
-- Create index "bronzegcpdataproccluster_location" to table: "gcp_dataproc_clusters"
CREATE INDEX "bronzegcpdataproccluster_location" ON "bronze"."gcp_dataproc_clusters" ("location");
-- Create index "bronzegcpdataproccluster_project_id" to table: "gcp_dataproc_clusters"
CREATE INDEX "bronzegcpdataproccluster_project_id" ON "bronze"."gcp_dataproc_clusters" ("project_id");
-- Create "gcp_containeranalysis_occurrences" table
CREATE TABLE "bronze"."gcp_containeranalysis_occurrences" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_uri" character varying NULL,
  "note_name" character varying NULL,
  "kind" bigint NULL,
  "remediation" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "vulnerability_json" jsonb NULL,
  "build_json" jsonb NULL,
  "image_json" jsonb NULL,
  "package_json" jsonb NULL,
  "deployment_json" jsonb NULL,
  "discovery_json" jsonb NULL,
  "attestation_json" jsonb NULL,
  "upgrade_json" jsonb NULL,
  "compliance_json" jsonb NULL,
  "dsse_attestation_json" jsonb NULL,
  "sbom_reference_json" jsonb NULL,
  "envelope_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcontaineranalysisoccurrence_collected_at" to table: "gcp_containeranalysis_occurrences"
CREATE INDEX "bronzegcpcontaineranalysisoccurrence_collected_at" ON "bronze"."gcp_containeranalysis_occurrences" ("collected_at");
-- Create index "bronzegcpcontaineranalysisoccurrence_kind" to table: "gcp_containeranalysis_occurrences"
CREATE INDEX "bronzegcpcontaineranalysisoccurrence_kind" ON "bronze"."gcp_containeranalysis_occurrences" ("kind");
-- Create index "bronzegcpcontaineranalysisoccurrence_note_name" to table: "gcp_containeranalysis_occurrences"
CREATE INDEX "bronzegcpcontaineranalysisoccurrence_note_name" ON "bronze"."gcp_containeranalysis_occurrences" ("note_name");
-- Create index "bronzegcpcontaineranalysisoccurrence_project_id" to table: "gcp_containeranalysis_occurrences"
CREATE INDEX "bronzegcpcontaineranalysisoccurrence_project_id" ON "bronze"."gcp_containeranalysis_occurrences" ("project_id");
-- Create index "bronzegcpcontaineranalysisoccurrence_resource_uri" to table: "gcp_containeranalysis_occurrences"
CREATE INDEX "bronzegcpcontaineranalysisoccurrence_resource_uri" ON "bronze"."gcp_containeranalysis_occurrences" ("resource_uri");
-- Create "gcp_containeranalysis_notes" table
CREATE TABLE "bronze"."gcp_containeranalysis_notes" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "short_description" character varying NULL,
  "long_description" character varying NULL,
  "kind" bigint NULL,
  "expiration_time" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "related_url_json" jsonb NULL,
  "related_note_names" jsonb NULL,
  "vulnerability_json" jsonb NULL,
  "build_json" jsonb NULL,
  "image_json" jsonb NULL,
  "package_json" jsonb NULL,
  "deployment_json" jsonb NULL,
  "discovery_json" jsonb NULL,
  "attestation_json" jsonb NULL,
  "upgrade_json" jsonb NULL,
  "compliance_json" jsonb NULL,
  "dsse_attestation_json" jsonb NULL,
  "vulnerability_assessment_json" jsonb NULL,
  "sbom_reference_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcontaineranalysisnote_collected_at" to table: "gcp_containeranalysis_notes"
CREATE INDEX "bronzegcpcontaineranalysisnote_collected_at" ON "bronze"."gcp_containeranalysis_notes" ("collected_at");
-- Create index "bronzegcpcontaineranalysisnote_kind" to table: "gcp_containeranalysis_notes"
CREATE INDEX "bronzegcpcontaineranalysisnote_kind" ON "bronze"."gcp_containeranalysis_notes" ("kind");
-- Create index "bronzegcpcontaineranalysisnote_project_id" to table: "gcp_containeranalysis_notes"
CREATE INDEX "bronzegcpcontaineranalysisnote_project_id" ON "bronze"."gcp_containeranalysis_notes" ("project_id");
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
-- Create "gcp_compute_interconnects" table
CREATE TABLE "bronze"."gcp_compute_interconnects" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "location" character varying NULL,
  "interconnect_type" character varying NULL,
  "link_type" character varying NULL,
  "admin_enabled" boolean NOT NULL DEFAULT false,
  "operational_status" character varying NULL,
  "provisioned_link_count" bigint NULL,
  "requested_link_count" bigint NULL,
  "peer_ip_address" character varying NULL,
  "google_ip_address" character varying NULL,
  "google_reference_id" character varying NULL,
  "noc_contact_email" character varying NULL,
  "customer_name" character varying NULL,
  "state" character varying NULL,
  "creation_timestamp" character varying NULL,
  "expected_outages_json" jsonb NULL,
  "circuit_infos_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeinterconnect_collected_at" to table: "gcp_compute_interconnects"
CREATE INDEX "bronzegcpcomputeinterconnect_collected_at" ON "bronze"."gcp_compute_interconnects" ("collected_at");
-- Create index "bronzegcpcomputeinterconnect_project_id" to table: "gcp_compute_interconnects"
CREATE INDEX "bronzegcpcomputeinterconnect_project_id" ON "bronze"."gcp_compute_interconnects" ("project_id");
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
-- Create "gcp_accesscontextmanager_access_levels" table
CREATE TABLE "bronze"."gcp_accesscontextmanager_access_levels" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "title" character varying NULL,
  "description" character varying NULL,
  "basic_json" jsonb NULL,
  "custom_json" jsonb NULL,
  "access_policy_name" character varying NOT NULL,
  "organization_id" character varying NOT NULL,
  "bronze_gcp_access_context_manager_access_policy_access_levels" character varying NOT NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_accesscontextmanager_access_levels_gcp_accesscontextmanager" FOREIGN KEY ("bronze_gcp_access_context_manager_access_policy_access_levels") REFERENCES "bronze"."gcp_accesscontextmanager_access_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "bronzegcpaccesscontextmanageraccesslevel_access_policy_name" to table: "gcp_accesscontextmanager_access_levels"
CREATE INDEX "bronzegcpaccesscontextmanageraccesslevel_access_policy_name" ON "bronze"."gcp_accesscontextmanager_access_levels" ("access_policy_name");
-- Create index "bronzegcpaccesscontextmanageraccesslevel_collected_at" to table: "gcp_accesscontextmanager_access_levels"
CREATE INDEX "bronzegcpaccesscontextmanageraccesslevel_collected_at" ON "bronze"."gcp_accesscontextmanager_access_levels" ("collected_at");
-- Create index "bronzegcpaccesscontextmanageraccesslevel_organization_id" to table: "gcp_accesscontextmanager_access_levels"
CREATE INDEX "bronzegcpaccesscontextmanageraccesslevel_organization_id" ON "bronze"."gcp_accesscontextmanager_access_levels" ("organization_id");
-- Create "gcp_accesscontextmanager_service_perimeters" table
CREATE TABLE "bronze"."gcp_accesscontextmanager_service_perimeters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "title" character varying NULL,
  "description" character varying NULL,
  "perimeter_type" bigint NOT NULL DEFAULT 0,
  "etag" character varying NULL,
  "use_explicit_dry_run_spec" boolean NOT NULL DEFAULT false,
  "status_json" jsonb NULL,
  "spec_json" jsonb NULL,
  "access_policy_name" character varying NOT NULL,
  "organization_id" character varying NOT NULL,
  "bronze_gcp_access_context_manager_access_policy_service_perimet" character varying NOT NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_accesscontextmanager_service_perimeters_gcp_accesscontextma" FOREIGN KEY ("bronze_gcp_access_context_manager_access_policy_service_perimet") REFERENCES "bronze"."gcp_accesscontextmanager_access_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "bronzegcpaccesscontextmanagerserviceperimeter_access_policy_nam" to table: "gcp_accesscontextmanager_service_perimeters"
CREATE INDEX "bronzegcpaccesscontextmanagerserviceperimeter_access_policy_nam" ON "bronze"."gcp_accesscontextmanager_service_perimeters" ("access_policy_name");
-- Create index "bronzegcpaccesscontextmanagerserviceperimeter_collected_at" to table: "gcp_accesscontextmanager_service_perimeters"
CREATE INDEX "bronzegcpaccesscontextmanagerserviceperimeter_collected_at" ON "bronze"."gcp_accesscontextmanager_service_perimeters" ("collected_at");
-- Create index "bronzegcpaccesscontextmanagerserviceperimeter_organization_id" to table: "gcp_accesscontextmanager_service_perimeters"
CREATE INDEX "bronzegcpaccesscontextmanagerserviceperimeter_organization_id" ON "bronze"."gcp_accesscontextmanager_service_perimeters" ("organization_id");
-- Create index "bronzegcpaccesscontextmanagerserviceperimeter_perimeter_type" to table: "gcp_accesscontextmanager_service_perimeters"
CREATE INDEX "bronzegcpaccesscontextmanagerserviceperimeter_perimeter_type" ON "bronze"."gcp_accesscontextmanager_service_perimeters" ("perimeter_type");
-- Create "gcp_appengine_applications" table
CREATE TABLE "bronze"."gcp_appengine_applications" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "auth_domain" character varying NULL,
  "location_id" character varying NULL,
  "code_bucket" character varying NULL,
  "default_cookie_expiration" character varying NULL,
  "serving_status" integer NULL,
  "default_hostname" character varying NULL,
  "default_bucket" character varying NULL,
  "gcr_domain" character varying NULL,
  "database_type" integer NULL,
  "feature_settings_json" jsonb NULL,
  "iap_json" jsonb NULL,
  "dispatch_rules_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpappengineapplication_collected_at" to table: "gcp_appengine_applications"
CREATE INDEX "bronzegcpappengineapplication_collected_at" ON "bronze"."gcp_appengine_applications" ("collected_at");
-- Create index "bronzegcpappengineapplication_project_id" to table: "gcp_appengine_applications"
CREATE INDEX "bronzegcpappengineapplication_project_id" ON "bronze"."gcp_appengine_applications" ("project_id");
-- Create "gcp_appengine_services" table
CREATE TABLE "bronze"."gcp_appengine_services" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "split_json" jsonb NULL,
  "labels_json" jsonb NULL,
  "network_settings_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  "bronze_gcp_app_engine_application_services" character varying NOT NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_appengine_services_gcp_appengine_applications_services" FOREIGN KEY ("bronze_gcp_app_engine_application_services") REFERENCES "bronze"."gcp_appengine_applications" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "bronzegcpappengineservice_collected_at" to table: "gcp_appengine_services"
CREATE INDEX "bronzegcpappengineservice_collected_at" ON "bronze"."gcp_appengine_services" ("collected_at");
-- Create index "bronzegcpappengineservice_project_id" to table: "gcp_appengine_services"
CREATE INDEX "bronzegcpappengineservice_project_id" ON "bronze"."gcp_appengine_services" ("project_id");
-- Create "gcp_bigtable_instances" table
CREATE TABLE "bronze"."gcp_bigtable_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "display_name" character varying NULL,
  "state" integer NULL,
  "instance_type" integer NULL,
  "create_time" character varying NULL,
  "satisfies_pzs" boolean NULL,
  "labels_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpbigtableinstance_collected_at" to table: "gcp_bigtable_instances"
CREATE INDEX "bronzegcpbigtableinstance_collected_at" ON "bronze"."gcp_bigtable_instances" ("collected_at");
-- Create index "bronzegcpbigtableinstance_project_id" to table: "gcp_bigtable_instances"
CREATE INDEX "bronzegcpbigtableinstance_project_id" ON "bronze"."gcp_bigtable_instances" ("project_id");
-- Create "gcp_bigtable_clusters" table
CREATE TABLE "bronze"."gcp_bigtable_clusters" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "location" character varying NULL,
  "state" integer NULL,
  "serve_nodes" integer NULL,
  "default_storage_type" integer NULL,
  "encryption_config_json" jsonb NULL,
  "cluster_config_json" jsonb NULL,
  "instance_name" character varying NOT NULL,
  "project_id" character varying NOT NULL,
  "bronze_gcp_bigtable_instance_clusters" character varying NOT NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_bigtable_clusters_gcp_bigtable_instances_clusters" FOREIGN KEY ("bronze_gcp_bigtable_instance_clusters") REFERENCES "bronze"."gcp_bigtable_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "bronzegcpbigtablecluster_collected_at" to table: "gcp_bigtable_clusters"
CREATE INDEX "bronzegcpbigtablecluster_collected_at" ON "bronze"."gcp_bigtable_clusters" ("collected_at");
-- Create index "bronzegcpbigtablecluster_instance_name" to table: "gcp_bigtable_clusters"
CREATE INDEX "bronzegcpbigtablecluster_instance_name" ON "bronze"."gcp_bigtable_clusters" ("instance_name");
-- Create index "bronzegcpbigtablecluster_project_id" to table: "gcp_bigtable_clusters"
CREATE INDEX "bronzegcpbigtablecluster_project_id" ON "bronze"."gcp_bigtable_clusters" ("project_id");
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
-- Create "gcp_compute_firewalls" table
CREATE TABLE "bronze"."gcp_compute_firewalls" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "self_link" character varying NULL,
  "creation_timestamp" character varying NULL,
  "network" character varying NULL,
  "priority" integer NOT NULL DEFAULT 1000,
  "direction" character varying NULL,
  "disabled" boolean NOT NULL DEFAULT false,
  "source_ranges_json" jsonb NULL,
  "destination_ranges_json" jsonb NULL,
  "source_tags_json" jsonb NULL,
  "target_tags_json" jsonb NULL,
  "source_service_accounts_json" jsonb NULL,
  "target_service_accounts_json" jsonb NULL,
  "log_config_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputefirewall_collected_at" to table: "gcp_compute_firewalls"
CREATE INDEX "bronzegcpcomputefirewall_collected_at" ON "bronze"."gcp_compute_firewalls" ("collected_at");
-- Create index "bronzegcpcomputefirewall_project_id" to table: "gcp_compute_firewalls"
CREATE INDEX "bronzegcpcomputefirewall_project_id" ON "bronze"."gcp_compute_firewalls" ("project_id");
-- Create "gcp_compute_firewall_alloweds" table
CREATE TABLE "bronze"."gcp_compute_firewall_alloweds" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "ip_protocol" character varying NOT NULL,
  "ports_json" jsonb NULL,
  "bronze_gcp_compute_firewall_allowed" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_firewall_alloweds_gcp_compute_firewalls_allowed" FOREIGN KEY ("bronze_gcp_compute_firewall_allowed") REFERENCES "bronze"."gcp_compute_firewalls" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_compute_firewall_denieds" table
CREATE TABLE "bronze"."gcp_compute_firewall_denieds" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "ip_protocol" character varying NOT NULL,
  "ports_json" jsonb NULL,
  "bronze_gcp_compute_firewall_denied" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_firewall_denieds_gcp_compute_firewalls_denied" FOREIGN KEY ("bronze_gcp_compute_firewall_denied") REFERENCES "bronze"."gcp_compute_firewalls" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "gcp_compute_project_metadata" table
CREATE TABLE "bronze"."gcp_compute_project_metadata" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "default_service_account" character varying NULL,
  "default_network_tier" character varying NULL,
  "xpn_project_status" character varying NULL,
  "creation_timestamp" character varying NULL,
  "usage_export_location_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpcomputeprojectmetadata_collected_at" to table: "gcp_compute_project_metadata"
CREATE INDEX "bronzegcpcomputeprojectmetadata_collected_at" ON "bronze"."gcp_compute_project_metadata" ("collected_at");
-- Create index "bronzegcpcomputeprojectmetadata_project_id" to table: "gcp_compute_project_metadata"
CREATE INDEX "bronzegcpcomputeprojectmetadata_project_id" ON "bronze"."gcp_compute_project_metadata" ("project_id");
-- Create "gcp_compute_project_metadata_items" table
CREATE TABLE "bronze"."gcp_compute_project_metadata_items" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" text NULL,
  "bronze_gcp_compute_project_metadata_items" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_compute_project_metadata_items_gcp_compute_project_metadata" FOREIGN KEY ("bronze_gcp_compute_project_metadata_items") REFERENCES "bronze"."gcp_compute_project_metadata" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "gcp_dns_managed_zones" table
CREATE TABLE "bronze"."gcp_dns_managed_zones" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "dns_name" character varying NULL,
  "description" character varying NULL,
  "visibility" character varying NULL,
  "creation_time" character varying NULL,
  "dnssec_config_json" jsonb NULL,
  "private_visibility_config_json" jsonb NULL,
  "forwarding_config_json" jsonb NULL,
  "peering_config_json" jsonb NULL,
  "cloud_logging_config_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpdnsmanagedzone_collected_at" to table: "gcp_dns_managed_zones"
CREATE INDEX "bronzegcpdnsmanagedzone_collected_at" ON "bronze"."gcp_dns_managed_zones" ("collected_at");
-- Create index "bronzegcpdnsmanagedzone_project_id" to table: "gcp_dns_managed_zones"
CREATE INDEX "bronzegcpdnsmanagedzone_project_id" ON "bronze"."gcp_dns_managed_zones" ("project_id");
-- Create "gcp_dns_managed_zone_labels" table
CREATE TABLE "bronze"."gcp_dns_managed_zone_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcpdns_managed_zone_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_dns_managed_zone_labels_gcp_dns_managed_zones_labels" FOREIGN KEY ("bronze_gcpdns_managed_zone_labels") REFERENCES "bronze"."gcp_dns_managed_zones" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_folder_iam_policies" table
CREATE TABLE "bronze"."gcp_folder_iam_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_name" character varying NOT NULL,
  "etag" character varying NULL,
  "version" bigint NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpfolderiampolicy_collected_at" to table: "gcp_folder_iam_policies"
CREATE INDEX "bronzegcpfolderiampolicy_collected_at" ON "bronze"."gcp_folder_iam_policies" ("collected_at");
-- Create "gcp_folder_iam_policy_bindings" table
CREATE TABLE "bronze"."gcp_folder_iam_policy_bindings" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "role" character varying NOT NULL,
  "members_json" jsonb NULL,
  "condition_json" jsonb NULL,
  "bronze_gcp_folder_iam_policy_bindings" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_folder_iam_policy_bindings_gcp_folder_iam_policies_bindings" FOREIGN KEY ("bronze_gcp_folder_iam_policy_bindings") REFERENCES "bronze"."gcp_folder_iam_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_folders" table
CREATE TABLE "bronze"."gcp_folders" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "display_name" character varying NULL,
  "state" character varying NULL,
  "parent" character varying NULL,
  "etag" character varying NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpfolder_collected_at" to table: "gcp_folders"
CREATE INDEX "bronzegcpfolder_collected_at" ON "bronze"."gcp_folders" ("collected_at");
-- Create index "bronzegcpfolder_parent" to table: "gcp_folders"
CREATE INDEX "bronzegcpfolder_parent" ON "bronze"."gcp_folders" ("parent");
-- Create index "bronzegcpfolder_state" to table: "gcp_folders"
CREATE INDEX "bronzegcpfolder_state" ON "bronze"."gcp_folders" ("state");
-- Create "gcp_folder_labels" table
CREATE TABLE "bronze"."gcp_folder_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_folder_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_folder_labels_gcp_folders_labels" FOREIGN KEY ("bronze_gcp_folder_labels") REFERENCES "bronze"."gcp_folders" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_org_iam_policies" table
CREATE TABLE "bronze"."gcp_org_iam_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_name" character varying NOT NULL,
  "etag" character varying NULL,
  "version" bigint NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcporgiampolicy_collected_at" to table: "gcp_org_iam_policies"
CREATE INDEX "bronzegcporgiampolicy_collected_at" ON "bronze"."gcp_org_iam_policies" ("collected_at");
-- Create "gcp_org_iam_policy_bindings" table
CREATE TABLE "bronze"."gcp_org_iam_policy_bindings" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "role" character varying NOT NULL,
  "members_json" jsonb NULL,
  "condition_json" jsonb NULL,
  "bronze_gcp_org_iam_policy_bindings" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_org_iam_policy_bindings_gcp_org_iam_policies_bindings" FOREIGN KEY ("bronze_gcp_org_iam_policy_bindings") REFERENCES "bronze"."gcp_org_iam_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_project_iam_policies" table
CREATE TABLE "bronze"."gcp_project_iam_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "resource_name" character varying NOT NULL,
  "etag" character varying NULL,
  "version" bigint NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpprojectiampolicy_collected_at" to table: "gcp_project_iam_policies"
CREATE INDEX "bronzegcpprojectiampolicy_collected_at" ON "bronze"."gcp_project_iam_policies" ("collected_at");
-- Create index "bronzegcpprojectiampolicy_project_id" to table: "gcp_project_iam_policies"
CREATE INDEX "bronzegcpprojectiampolicy_project_id" ON "bronze"."gcp_project_iam_policies" ("project_id");
-- Create "gcp_project_iam_policy_bindings" table
CREATE TABLE "bronze"."gcp_project_iam_policy_bindings" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "role" character varying NOT NULL,
  "members_json" jsonb NULL,
  "condition_json" jsonb NULL,
  "bronze_gcp_project_iam_policy_bindings" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_project_iam_policy_bindings_gcp_project_iam_policies_bindin" FOREIGN KEY ("bronze_gcp_project_iam_policy_bindings") REFERENCES "bronze"."gcp_project_iam_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
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
-- Create "gcp_run_services" table
CREATE TABLE "bronze"."gcp_run_services" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "description" character varying NULL,
  "uid" character varying NULL,
  "generation" bigint NULL,
  "labels_json" jsonb NULL,
  "annotations_json" jsonb NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  "creator" character varying NULL,
  "last_modifier" character varying NULL,
  "ingress" bigint NULL,
  "launch_stage" bigint NULL,
  "template_json" jsonb NULL,
  "traffic_json" jsonb NULL,
  "uri" character varying NULL,
  "observed_generation" bigint NULL,
  "terminal_condition_json" jsonb NULL,
  "conditions_json" jsonb NULL,
  "latest_ready_revision" character varying NULL,
  "latest_created_revision" character varying NULL,
  "traffic_statuses_json" jsonb NULL,
  "reconciling" boolean NOT NULL DEFAULT false,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcprunservice_collected_at" to table: "gcp_run_services"
CREATE INDEX "bronzegcprunservice_collected_at" ON "bronze"."gcp_run_services" ("collected_at");
-- Create index "bronzegcprunservice_location" to table: "gcp_run_services"
CREATE INDEX "bronzegcprunservice_location" ON "bronze"."gcp_run_services" ("location");
-- Create index "bronzegcprunservice_project_id" to table: "gcp_run_services"
CREATE INDEX "bronzegcprunservice_project_id" ON "bronze"."gcp_run_services" ("project_id");
-- Create "gcp_run_revisions" table
CREATE TABLE "bronze"."gcp_run_revisions" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "uid" character varying NULL,
  "generation" bigint NULL,
  "labels_json" jsonb NULL,
  "annotations_json" jsonb NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "delete_time" character varying NULL,
  "launch_stage" bigint NULL,
  "service_name" character varying NULL,
  "scaling_json" jsonb NULL,
  "containers_json" jsonb NULL,
  "volumes_json" jsonb NULL,
  "execution_environment" bigint NULL,
  "encryption_key" character varying NULL,
  "max_instance_request_concurrency" bigint NULL,
  "timeout" character varying NULL,
  "service_account" character varying NULL,
  "reconciling" boolean NOT NULL DEFAULT false,
  "conditions_json" jsonb NULL,
  "observed_generation" bigint NULL,
  "log_uri" character varying NULL,
  "etag" character varying NULL,
  "project_id" character varying NOT NULL,
  "location" character varying NOT NULL,
  "bronze_gcp_run_service_revisions" character varying NOT NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_run_revisions_gcp_run_services_revisions" FOREIGN KEY ("bronze_gcp_run_service_revisions") REFERENCES "bronze"."gcp_run_services" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "bronzegcprunrevision_collected_at" to table: "gcp_run_revisions"
CREATE INDEX "bronzegcprunrevision_collected_at" ON "bronze"."gcp_run_revisions" ("collected_at");
-- Create index "bronzegcprunrevision_location" to table: "gcp_run_revisions"
CREATE INDEX "bronzegcprunrevision_location" ON "bronze"."gcp_run_revisions" ("location");
-- Create index "bronzegcprunrevision_project_id" to table: "gcp_run_revisions"
CREATE INDEX "bronzegcprunrevision_project_id" ON "bronze"."gcp_run_revisions" ("project_id");
-- Create index "bronzegcprunrevision_service_name" to table: "gcp_run_revisions"
CREATE INDEX "bronzegcprunrevision_service_name" ON "bronze"."gcp_run_revisions" ("service_name");
-- Create "gcp_secretmanager_secrets" table
CREATE TABLE "bronze"."gcp_secretmanager_secrets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "create_time" character varying NULL,
  "etag" character varying NULL,
  "replication_json" jsonb NULL,
  "rotation_json" jsonb NULL,
  "topics_json" jsonb NULL,
  "version_aliases_json" jsonb NULL,
  "annotations_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpsecretmanagersecret_collected_at" to table: "gcp_secretmanager_secrets"
CREATE INDEX "bronzegcpsecretmanagersecret_collected_at" ON "bronze"."gcp_secretmanager_secrets" ("collected_at");
-- Create index "bronzegcpsecretmanagersecret_project_id" to table: "gcp_secretmanager_secrets"
CREATE INDEX "bronzegcpsecretmanagersecret_project_id" ON "bronze"."gcp_secretmanager_secrets" ("project_id");
-- Create "gcp_secretmanager_secret_labels" table
CREATE TABLE "bronze"."gcp_secretmanager_secret_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_secret_manager_secret_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_secretmanager_secret_labels_gcp_secretmanager_secrets_label" FOREIGN KEY ("bronze_gcp_secret_manager_secret_labels") REFERENCES "bronze"."gcp_secretmanager_secrets" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_spanner_instances" table
CREATE TABLE "bronze"."gcp_spanner_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "config" character varying NULL,
  "display_name" character varying NULL,
  "node_count" integer NULL,
  "processing_units" integer NULL,
  "state" bigint NULL,
  "labels_json" jsonb NULL,
  "endpoint_uris_json" jsonb NULL,
  "create_time" character varying NULL,
  "update_time" character varying NULL,
  "edition" bigint NULL,
  "default_backup_schedule_type" bigint NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpspannerinstance_collected_at" to table: "gcp_spanner_instances"
CREATE INDEX "bronzegcpspannerinstance_collected_at" ON "bronze"."gcp_spanner_instances" ("collected_at");
-- Create index "bronzegcpspannerinstance_project_id" to table: "gcp_spanner_instances"
CREATE INDEX "bronzegcpspannerinstance_project_id" ON "bronze"."gcp_spanner_instances" ("project_id");
-- Create "gcp_spanner_databases" table
CREATE TABLE "bronze"."gcp_spanner_databases" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "state" bigint NULL,
  "create_time" character varying NULL,
  "restore_info_json" jsonb NULL,
  "encryption_config_json" jsonb NULL,
  "encryption_info_json" jsonb NULL,
  "version_retention_period" character varying NULL,
  "earliest_version_time" character varying NULL,
  "default_leader" character varying NULL,
  "database_dialect" bigint NULL,
  "enable_drop_protection" boolean NOT NULL DEFAULT false,
  "reconciling" boolean NOT NULL DEFAULT false,
  "instance_name" character varying NULL,
  "project_id" character varying NOT NULL,
  "bronze_gcp_spanner_instance_databases" character varying NULL,
  PRIMARY KEY ("resource_id"),
  CONSTRAINT "gcp_spanner_databases_gcp_spanner_instances_databases" FOREIGN KEY ("bronze_gcp_spanner_instance_databases") REFERENCES "bronze"."gcp_spanner_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE SET NULL
);
-- Create index "bronzegcpspannerdatabase_collected_at" to table: "gcp_spanner_databases"
CREATE INDEX "bronzegcpspannerdatabase_collected_at" ON "bronze"."gcp_spanner_databases" ("collected_at");
-- Create index "bronzegcpspannerdatabase_project_id" to table: "gcp_spanner_databases"
CREATE INDEX "bronzegcpspannerdatabase_project_id" ON "bronze"."gcp_spanner_databases" ("project_id");
-- Create "gcp_sql_instances" table
CREATE TABLE "bronze"."gcp_sql_instances" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "database_version" character varying NULL,
  "state" character varying NULL,
  "region" character varying NULL,
  "gce_zone" character varying NULL,
  "secondary_gce_zone" character varying NULL,
  "instance_type" character varying NULL,
  "connection_name" character varying NULL,
  "service_account_email_address" character varying NULL,
  "self_link" character varying NULL,
  "settings_json" jsonb NULL,
  "server_ca_cert_json" jsonb NULL,
  "ip_addresses_json" jsonb NULL,
  "replica_configuration_json" jsonb NULL,
  "failover_replica_json" jsonb NULL,
  "disk_encryption_configuration_json" jsonb NULL,
  "disk_encryption_status_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpsqlinstance_collected_at" to table: "gcp_sql_instances"
CREATE INDEX "bronzegcpsqlinstance_collected_at" ON "bronze"."gcp_sql_instances" ("collected_at");
-- Create index "bronzegcpsqlinstance_project_id" to table: "gcp_sql_instances"
CREATE INDEX "bronzegcpsqlinstance_project_id" ON "bronze"."gcp_sql_instances" ("project_id");
-- Create index "bronzegcpsqlinstance_state" to table: "gcp_sql_instances"
CREATE INDEX "bronzegcpsqlinstance_state" ON "bronze"."gcp_sql_instances" ("state");
-- Create "gcp_sql_instance_labels" table
CREATE TABLE "bronze"."gcp_sql_instance_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NULL,
  "bronze_gcpsql_instance_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_sql_instance_labels_gcp_sql_instances_labels" FOREIGN KEY ("bronze_gcpsql_instance_labels") REFERENCES "bronze"."gcp_sql_instances" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_storage_bucket_iam_policies" table
CREATE TABLE "bronze"."gcp_storage_bucket_iam_policies" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "bucket_name" character varying NOT NULL,
  "etag" character varying NULL,
  "version" bigint NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpstoragebucketiampolicy_collected_at" to table: "gcp_storage_bucket_iam_policies"
CREATE INDEX "bronzegcpstoragebucketiampolicy_collected_at" ON "bronze"."gcp_storage_bucket_iam_policies" ("collected_at");
-- Create index "bronzegcpstoragebucketiampolicy_project_id" to table: "gcp_storage_bucket_iam_policies"
CREATE INDEX "bronzegcpstoragebucketiampolicy_project_id" ON "bronze"."gcp_storage_bucket_iam_policies" ("project_id");
-- Create "gcp_storage_bucket_iam_policy_bindings" table
CREATE TABLE "bronze"."gcp_storage_bucket_iam_policy_bindings" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "role" character varying NOT NULL,
  "members_json" jsonb NULL,
  "condition_json" jsonb NULL,
  "bronze_gcp_storage_bucket_iam_policy_bindings" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_storage_bucket_iam_policy_bindings_gcp_storage_bucket_iam_p" FOREIGN KEY ("bronze_gcp_storage_bucket_iam_policy_bindings") REFERENCES "bronze"."gcp_storage_bucket_iam_policies" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create "gcp_storage_buckets" table
CREATE TABLE "bronze"."gcp_storage_buckets" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "name" character varying NOT NULL,
  "location" character varying NULL,
  "storage_class" character varying NULL,
  "project_number" character varying NULL,
  "time_created" character varying NULL,
  "updated" character varying NULL,
  "default_event_based_hold" boolean NOT NULL DEFAULT false,
  "metageneration" character varying NULL,
  "etag" character varying NULL,
  "iam_configuration_json" jsonb NULL,
  "encryption_json" jsonb NULL,
  "lifecycle_json" jsonb NULL,
  "versioning_json" jsonb NULL,
  "retention_policy_json" jsonb NULL,
  "logging_json" jsonb NULL,
  "cors_json" jsonb NULL,
  "website_json" jsonb NULL,
  "autoclass_json" jsonb NULL,
  "project_id" character varying NOT NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzegcpstoragebucket_collected_at" to table: "gcp_storage_buckets"
CREATE INDEX "bronzegcpstoragebucket_collected_at" ON "bronze"."gcp_storage_buckets" ("collected_at");
-- Create index "bronzegcpstoragebucket_project_id" to table: "gcp_storage_buckets"
CREATE INDEX "bronzegcpstoragebucket_project_id" ON "bronze"."gcp_storage_buckets" ("project_id");
-- Create "gcp_storage_bucket_labels" table
CREATE TABLE "bronze"."gcp_storage_bucket_labels" (
  "id" bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
  "key" character varying NOT NULL,
  "value" character varying NOT NULL,
  "bronze_gcp_storage_bucket_labels" character varying NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "gcp_storage_bucket_labels_gcp_storage_buckets_labels" FOREIGN KEY ("bronze_gcp_storage_bucket_labels") REFERENCES "bronze"."gcp_storage_buckets" ("resource_id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
