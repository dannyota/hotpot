package gcp

import (
	"database/sql"

	"entgo.io/ent/dialect"

	"danny.vn/hotpot/pkg/admin"
	"danny.vn/hotpot/pkg/admin/bronze/gcp/compute"
	lh "danny.vn/hotpot/pkg/admin/listhandler"
)

// Register registers all GCP admin routes.
func Register(driver dialect.Driver, db *sql.DB) {
	compute.Register(driver, db)
	lh.RegisterSQL(db, sqlTables)
}

func bronzeGCP(api, table, label, group string) lh.SQLTable {
	return lh.SQLTable{
		API:    "/api/v1/bronze/gcp/" + api,
		Schema: "bronze",
		Table:  table,
		Nav:    admin.NavMeta{Label: label, Group: []string{"Bronze", "GCP", group}},
	}
}

var sqlTables = []lh.SQLTable{
	// Access Context Manager
	bronzeGCP("accesscontextmanager/access-policies", "gcp_accesscontextmanager_access_policies", "Access Policies", "Access Context Manager"),
	bronzeGCP("accesscontextmanager/access-levels", "gcp_accesscontextmanager_access_levels", "Access Levels", "Access Context Manager"),
	bronzeGCP("accesscontextmanager/service-perimeters", "gcp_accesscontextmanager_service_perimeters", "Service Perimeters", "Access Context Manager"),

	// AlloyDB
	bronzeGCP("alloydb/clusters", "gcp_alloydb_clusters", "Clusters", "AlloyDB"),

	// App Engine
	bronzeGCP("appengine/applications", "gcp_appengine_applications", "Applications", "App Engine"),
	bronzeGCP("appengine/services", "gcp_appengine_services", "Services", "App Engine"),

	// BigQuery
	bronzeGCP("bigquery/datasets", "gcp_bigquery_datasets", "Datasets", "BigQuery"),
	bronzeGCP("bigquery/tables", "gcp_bigquery_tables", "Tables", "BigQuery"),

	// Bigtable
	bronzeGCP("bigtable/instances", "gcp_bigtable_instances", "Instances", "Bigtable"),
	bronzeGCP("bigtable/clusters", "gcp_bigtable_clusters", "Clusters", "Bigtable"),

	// Binary Authorization
	bronzeGCP("binaryauthorization/policies", "gcp_binaryauthorization_policies", "Policies", "Binary Authorization"),
	bronzeGCP("binaryauthorization/attestors", "gcp_binaryauthorization_attestors", "Attestors", "Binary Authorization"),

	// Cloud Asset
	bronzeGCP("cloudasset/assets", "gcp_cloudasset_assets", "Assets", "Cloud Asset"),
	bronzeGCP("cloudasset/iam-policy-searches", "gcp_cloudasset_iam_policy_searches", "IAM Policy Searches", "Cloud Asset"),
	bronzeGCP("cloudasset/resource-searches", "gcp_cloudasset_resource_searches", "Resource Searches", "Cloud Asset"),

	// Cloud Functions
	bronzeGCP("cloudfunctions/functions", "gcp_cloudfunctions_functions", "Functions", "Cloud Functions"),

	// Container (GKE)
	{
		API: "/api/v1/bronze/gcp/container/clusters", Schema: "bronze",
		Table: "gcp_container_clusters", Nav: admin.NavMeta{Label: "Clusters", Group: []string{"Bronze", "GCP", "Container"}},
		Columns:             []string{"resource_id", "name", "location", "status", "current_master_version", "current_node_version", "current_node_count", "network", "subnetwork", "endpoint", "project_id", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "location", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "location", "project_id"},
	},

	// Container Analysis
	bronzeGCP("containeranalysis/notes", "gcp_containeranalysis_notes", "Notes", "Container Analysis"),
	bronzeGCP("containeranalysis/occurrences", "gcp_containeranalysis_occurrences", "Occurrences", "Container Analysis"),

	// Dataproc
	bronzeGCP("dataproc/clusters", "gcp_dataproc_clusters", "Clusters", "Dataproc"),

	// DNS
	bronzeGCP("dns/managed-zones", "gcp_dns_managed_zones", "Managed Zones", "DNS"),
	bronzeGCP("dns/policies", "gcp_dns_policies", "Policies", "DNS"),

	// Filestore
	bronzeGCP("filestore/instances", "gcp_filestore_instances", "Instances", "Filestore"),

	// IAM
	{
		API: "/api/v1/bronze/gcp/iam/service-accounts", Schema: "bronze",
		Table: "gcp_iam_service_accounts", Nav: admin.NavMeta{Label: "Service Accounts", Group: []string{"Bronze", "GCP", "IAM"}},
		Columns:             []string{"resource_id", "name", "email", "display_name", "disabled", "description", "project_id", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "disabled", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"disabled", "project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/iam/service-account-keys", Schema: "bronze",
		Table: "gcp_iam_service_account_keys", Nav: admin.NavMeta{Label: "Service Account Keys", Group: []string{"Bronze", "GCP", "IAM"}},
		Columns:             []string{"resource_id", "name", "service_account_email", "key_origin", "key_type", "key_algorithm", "valid_after_time", "valid_before_time", "disabled", "project_id", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "key_type", Kind: lh.Multi}, {Column: "key_origin", Kind: lh.Multi}, {Column: "disabled", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"key_type", "key_origin", "disabled", "project_id"},
	},

	// IAP
	bronzeGCP("iap/iam-policies", "gcp_iap_iam_policies", "IAM Policies", "IAP"),
	bronzeGCP("iap/settings", "gcp_iap_settings", "Settings", "IAP"),

	// KMS
	bronzeGCP("kms/key-rings", "gcp_kms_key_rings", "Key Rings", "KMS"),
	bronzeGCP("kms/crypto-keys", "gcp_kms_crypto_keys", "Crypto Keys", "KMS"),

	// Logging
	bronzeGCP("logging/buckets", "gcp_logging_buckets", "Buckets", "Logging"),
	bronzeGCP("logging/sinks", "gcp_logging_sinks", "Sinks", "Logging"),
	bronzeGCP("logging/log-exclusions", "gcp_logging_log_exclusions", "Log Exclusions", "Logging"),
	bronzeGCP("logging/log-metrics", "gcp_logging_log_metrics", "Log Metrics", "Logging"),

	// Monitoring
	bronzeGCP("monitoring/alert-policies", "gcp_monitoring_alert_policies", "Alert Policies", "Monitoring"),
	bronzeGCP("monitoring/uptime-checks", "gcp_monitoring_uptime_check_configs", "Uptime Checks", "Monitoring"),

	// Org Policy
	bronzeGCP("orgpolicy/policies", "gcp_orgpolicy_policies", "Policies", "Org Policy"),
	bronzeGCP("orgpolicy/constraints", "gcp_orgpolicy_constraints", "Constraints", "Org Policy"),
	bronzeGCP("orgpolicy/custom-constraints", "gcp_orgpolicy_custom_constraints", "Custom Constraints", "Org Policy"),

	// Pub/Sub
	bronzeGCP("pubsub/topics", "gcp_pubsub_topics", "Topics", "Pub/Sub"),
	bronzeGCP("pubsub/subscriptions", "gcp_pubsub_subscriptions", "Subscriptions", "Pub/Sub"),

	// Redis
	bronzeGCP("redis/instances", "gcp_redis_instances", "Instances", "Redis"),

	// Resource Manager
	bronzeGCP("resourcemanager/organizations", "gcp_organizations", "Organizations", "Resource Manager"),
	bronzeGCP("resourcemanager/folders", "gcp_folders", "Folders", "Resource Manager"),
	bronzeGCP("resourcemanager/projects", "gcp_projects", "Projects", "Resource Manager"),

	// Cloud Run
	bronzeGCP("run/services", "gcp_run_services", "Services", "Cloud Run"),
	bronzeGCP("run/revisions", "gcp_run_revisions", "Revisions", "Cloud Run"),

	// Secret Manager
	bronzeGCP("secretmanager/secrets", "gcp_secretmanager_secrets", "Secrets", "Secret Manager"),

	// Security Center
	bronzeGCP("securitycenter/findings", "gcp_securitycenter_findings", "Findings", "Security Center"),
	bronzeGCP("securitycenter/sources", "gcp_securitycenter_sources", "Sources", "Security Center"),
	bronzeGCP("securitycenter/notification-configs", "gcp_securitycenter_notification_configs", "Notification Configs", "Security Center"),

	// Service Usage
	bronzeGCP("serviceusage/enabled-services", "gcp_serviceusage_enabled_services", "Enabled Services", "Service Usage"),

	// Spanner
	bronzeGCP("spanner/instances", "gcp_spanner_instances", "Instances", "Spanner"),
	bronzeGCP("spanner/databases", "gcp_spanner_databases", "Databases", "Spanner"),

	// Cloud SQL
	{
		API: "/api/v1/bronze/gcp/sql/instances", Schema: "bronze",
		Table: "gcp_sql_instances", Nav: admin.NavMeta{Label: "Instances", Group: []string{"Bronze", "GCP", "Cloud SQL"}},
		Columns:             []string{"resource_id", "name", "database_version", "state", "region", "gce_zone", "instance_type", "connection_name", "project_id", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "state", Kind: lh.Multi}, {Column: "database_version", Kind: lh.Multi}, {Column: "region", Kind: lh.Multi}, {Column: "instance_type", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"state", "database_version", "region", "instance_type", "project_id"},
	},

	// Storage
	{
		API: "/api/v1/bronze/gcp/storage/buckets", Schema: "bronze",
		Table: "gcp_storage_buckets", Nav: admin.NavMeta{Label: "Buckets", Group: []string{"Bronze", "GCP", "Storage"}},
		Columns:             []string{"resource_id", "name", "location", "storage_class", "time_created", "default_event_based_hold", "project_id", "collected_at", "first_collected_at"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "location", Kind: lh.Multi}, {Column: "storage_class", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "collected_at", DefaultDesc: true,
		FilterOptionColumns: []string{"location", "storage_class", "project_id"},
	},

	// VPC Access
	bronzeGCP("vpcaccess/connectors", "gcp_vpc_access_connectors", "Connectors", "VPC Access"),

	// Container (GKE) — Child Tables
	{
		API: "/api/v1/bronze/gcp/container/node-pools", Schema: "bronze",
		Table: "gcp_container_cluster_node_pools", Nav: admin.NavMeta{Label: "Node Pools", Group: []string{"Bronze", "GCP", "Container"}},
		From: `SELECT np."id", np."name", np."version", np."status", np."status_message", np."initial_node_count", np."pod_ipv4_cidr_size",
			c."name" AS cluster_name, c."location" AS cluster_location, c."project_id"
			FROM "bronze"."gcp_container_cluster_node_pools" np
			LEFT JOIN "bronze"."gcp_container_clusters" c ON np."bronze_gcp_container_cluster_node_pools" = c."resource_id"`,
		Columns:             []string{"id", "name", "version", "status", "initial_node_count", "pod_ipv4_cidr_size", "cluster_name", "cluster_location", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "name", Kind: lh.Search}, {Column: "status", Kind: lh.Multi}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"status", "project_id"},
	},

	// IAM Policy Bindings
	{
		API: "/api/v1/bronze/gcp/resourcemanager/project-iam-bindings", Schema: "bronze",
		Table: "gcp_project_iam_policy_bindings", Nav: admin.NavMeta{Label: "Project IAM Bindings", Group: []string{"Bronze", "GCP", "Resource Manager"}},
		From: `SELECT b."id", b."role", b."members_json", b."condition_json",
			p."project_id"
			FROM "bronze"."gcp_project_iam_policy_bindings" b
			LEFT JOIN "bronze"."gcp_project_iam_policies" p ON b."bronze_gcp_project_iam_policy_bindings" = p."resource_id"`,
		Columns:             []string{"id", "role", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "role", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
	{
		API: "/api/v1/bronze/gcp/resourcemanager/folder-iam-bindings", Schema: "bronze",
		Table: "gcp_folder_iam_policy_bindings", Nav: admin.NavMeta{Label: "Folder IAM Bindings", Group: []string{"Bronze", "GCP", "Resource Manager"}},
		From: `SELECT b."id", b."role", b."members_json", b."condition_json",
			p."resource_name" AS folder_name
			FROM "bronze"."gcp_folder_iam_policy_bindings" b
			LEFT JOIN "bronze"."gcp_folder_iam_policies" p ON b."bronze_gcp_folder_iam_policy_bindings" = p."resource_id"`,
		Columns:             []string{"id", "role", "folder_name"},
		Filters:             []lh.SQLFilterDef{{Column: "role", Kind: lh.Search}},
		DefaultSort:         "id", DefaultDesc: true,
	},
	{
		API: "/api/v1/bronze/gcp/resourcemanager/org-iam-bindings", Schema: "bronze",
		Table: "gcp_org_iam_policy_bindings", Nav: admin.NavMeta{Label: "Org IAM Bindings", Group: []string{"Bronze", "GCP", "Resource Manager"}},
		From: `SELECT b."id", b."role", b."members_json", b."condition_json",
			p."resource_name" AS org_name
			FROM "bronze"."gcp_org_iam_policy_bindings" b
			LEFT JOIN "bronze"."gcp_org_iam_policies" p ON b."bronze_gcp_org_iam_policy_bindings" = p."resource_id"`,
		Columns:             []string{"id", "role", "org_name"},
		Filters:             []lh.SQLFilterDef{{Column: "role", Kind: lh.Search}},
		DefaultSort:         "id", DefaultDesc: true,
	},

	// Storage — Child Tables
	{
		API: "/api/v1/bronze/gcp/storage/bucket-iam-bindings", Schema: "bronze",
		Table: "gcp_storage_bucket_iam_policy_bindings", Nav: admin.NavMeta{Label: "Bucket IAM Bindings", Group: []string{"Bronze", "GCP", "Storage"}},
		From: `SELECT b."id", b."role", b."members_json", b."condition_json",
			p."bucket_name", p."project_id"
			FROM "bronze"."gcp_storage_bucket_iam_policy_bindings" b
			LEFT JOIN "bronze"."gcp_storage_bucket_iam_policies" p ON b."bronze_gcp_storage_bucket_iam_policy_bindings" = p."resource_id"`,
		Columns:             []string{"id", "role", "bucket_name", "project_id"},
		Filters:             []lh.SQLFilterDef{{Column: "role", Kind: lh.Search}, {Column: "project_id", Kind: lh.Multi}},
		DefaultSort:         "id", DefaultDesc: true,
		FilterOptionColumns: []string{"project_id"},
	},
}
