-- Add new schema named "bronze"
CREATE SCHEMA IF NOT EXISTS "bronze";
-- Create "apicatalog_endpoints_raw" table
CREATE TABLE "bronze"."apicatalog_endpoints_raw" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "log_source_id" character varying NULL,
  "name" character varying NULL,
  "service_name" character varying NULL,
  "upstream" character varying NULL,
  "uri" character varying NOT NULL,
  "method" character varying NOT NULL DEFAULT '',
  "route_status" character varying NOT NULL DEFAULT '',
  "plugin_auth" character varying NULL,
  "plugin_auth_enable" character varying NULL,
  "source_file" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzeapicatalogendpointsraw_log_source_id" to table: "apicatalog_endpoints_raw"
CREATE INDEX "bronzeapicatalogendpointsraw_log_source_id" ON "bronze"."apicatalog_endpoints_raw" ("log_source_id");
-- Create index "bronzeapicatalogendpointsraw_name_upstream_uri_method_route_sta" to table: "apicatalog_endpoints_raw"
CREATE UNIQUE INDEX "bronzeapicatalogendpointsraw_name_upstream_uri_method_route_sta" ON "bronze"."apicatalog_endpoints_raw" ("name", "upstream", "uri", "method", "route_status");
-- Create index "bronzeapicatalogendpointsraw_upstream" to table: "apicatalog_endpoints_raw"
CREATE INDEX "bronzeapicatalogendpointsraw_upstream" ON "bronze"."apicatalog_endpoints_raw" ("upstream");
