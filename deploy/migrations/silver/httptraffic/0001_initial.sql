-- Add new schema named "silver"
CREATE SCHEMA IF NOT EXISTS "silver";
-- Create "httptraffic_client_ip_5m" table
CREATE TABLE "silver"."httptraffic_client_ip_5m" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "endpoint_id" character varying NULL,
  "source_id" character varying NOT NULL,
  "window_start" timestamptz NOT NULL,
  "window_end" timestamptz NOT NULL,
  "uri" character varying NOT NULL,
  "method" character varying NULL,
  "client_ip" character varying NOT NULL,
  "country_code" character varying NULL,
  "country_name" character varying NULL,
  "asn" bigint NULL,
  "org_name" character varying NULL,
  "as_domain" character varying NULL,
  "asn_type" character varying NULL,
  "is_internal" boolean NOT NULL DEFAULT false,
  "request_count" bigint NOT NULL,
  "is_mapped" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("resource_id")
);
-- Create index "silverhttptrafficclientip5m_asn_window_start" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_asn_window_start" ON "silver"."httptraffic_client_ip_5m" ("asn", "window_start");
-- Create index "silverhttptrafficclientip5m_client_ip_window_start" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_client_ip_window_start" ON "silver"."httptraffic_client_ip_5m" ("client_ip", "window_start");
-- Create index "silverhttptrafficclientip5m_country_code_window_start" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_country_code_window_start" ON "silver"."httptraffic_client_ip_5m" ("country_code", "window_start");
-- Create index "silverhttptrafficclientip5m_endpoint_id_window_start" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_endpoint_id_window_start" ON "silver"."httptraffic_client_ip_5m" ("endpoint_id", "window_start");
-- Create index "silverhttptrafficclientip5m_is_internal" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_is_internal" ON "silver"."httptraffic_client_ip_5m" ("is_internal");
-- Create index "silverhttptrafficclientip5m_window_start" to table: "httptraffic_client_ip_5m"
CREATE INDEX "silverhttptrafficclientip5m_window_start" ON "silver"."httptraffic_client_ip_5m" ("window_start");
-- Create "httptraffic_traffic_5m" table
CREATE TABLE "silver"."httptraffic_traffic_5m" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "endpoint_id" character varying NULL,
  "source_id" character varying NOT NULL,
  "window_start" timestamptz NOT NULL,
  "window_end" timestamptz NOT NULL,
  "uri" character varying NOT NULL,
  "method" character varying NULL,
  "status_code" bigint NOT NULL,
  "request_count" bigint NOT NULL,
  "avg_request_time" double precision NULL,
  "max_request_time" double precision NULL,
  "total_body_bytes_sent" bigint NOT NULL DEFAULT 0,
  "unique_client_count" bigint NOT NULL DEFAULT 0,
  "access_level" character varying NULL,
  "service" character varying NULL,
  "is_mapped" boolean NOT NULL DEFAULT false,
  "is_method_mismatch" boolean NOT NULL DEFAULT false,
  "is_scanner_detected" boolean NOT NULL DEFAULT false,
  "is_lfi_detected" boolean NOT NULL DEFAULT false,
  "is_sqli_detected" boolean NOT NULL DEFAULT false,
  "is_rce_detected" boolean NOT NULL DEFAULT false,
  "is_xss_detected" boolean NOT NULL DEFAULT false,
  "is_ssrf_detected" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("resource_id")
);
-- Create index "silverhttptraffictraffic5m_endpoint_id_window_start" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_endpoint_id_window_start" ON "silver"."httptraffic_traffic_5m" ("endpoint_id", "window_start");
-- Create index "silverhttptraffictraffic5m_is_lfi_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_lfi_detected" ON "silver"."httptraffic_traffic_5m" ("is_lfi_detected");
-- Create index "silverhttptraffictraffic5m_is_mapped" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_mapped" ON "silver"."httptraffic_traffic_5m" ("is_mapped");
-- Create index "silverhttptraffictraffic5m_is_method_mismatch" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_method_mismatch" ON "silver"."httptraffic_traffic_5m" ("is_method_mismatch");
-- Create index "silverhttptraffictraffic5m_is_rce_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_rce_detected" ON "silver"."httptraffic_traffic_5m" ("is_rce_detected");
-- Create index "silverhttptraffictraffic5m_is_scanner_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_scanner_detected" ON "silver"."httptraffic_traffic_5m" ("is_scanner_detected");
-- Create index "silverhttptraffictraffic5m_is_sqli_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_sqli_detected" ON "silver"."httptraffic_traffic_5m" ("is_sqli_detected");
-- Create index "silverhttptraffictraffic5m_is_ssrf_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_ssrf_detected" ON "silver"."httptraffic_traffic_5m" ("is_ssrf_detected");
-- Create index "silverhttptraffictraffic5m_is_xss_detected" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_is_xss_detected" ON "silver"."httptraffic_traffic_5m" ("is_xss_detected");
-- Create index "silverhttptraffictraffic5m_source_id_window_start" to table: "httptraffic_traffic_5m"
CREATE INDEX "silverhttptraffictraffic5m_source_id_window_start" ON "silver"."httptraffic_traffic_5m" ("source_id", "window_start");
-- Create "httptraffic_user_agent_5m" table
CREATE TABLE "silver"."httptraffic_user_agent_5m" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "normalized_at" timestamptz NOT NULL,
  "endpoint_id" character varying NULL,
  "source_id" character varying NOT NULL,
  "window_start" timestamptz NOT NULL,
  "window_end" timestamptz NOT NULL,
  "uri" character varying NOT NULL,
  "method" character varying NULL,
  "user_agent" character varying NOT NULL,
  "ua_family" character varying NULL,
  "request_count" bigint NOT NULL,
  "is_mapped" boolean NOT NULL DEFAULT false,
  PRIMARY KEY ("resource_id")
);
-- Create index "silverhttptrafficuseragent5m_endpoint_id_window_start" to table: "httptraffic_user_agent_5m"
CREATE INDEX "silverhttptrafficuseragent5m_endpoint_id_window_start" ON "silver"."httptraffic_user_agent_5m" ("endpoint_id", "window_start");
-- Create index "silverhttptrafficuseragent5m_ua_family_window_start" to table: "httptraffic_user_agent_5m"
CREATE INDEX "silverhttptrafficuseragent5m_ua_family_window_start" ON "silver"."httptraffic_user_agent_5m" ("ua_family", "window_start");
-- Create index "silverhttptrafficuseragent5m_window_start" to table: "httptraffic_user_agent_5m"
CREATE INDEX "silverhttptrafficuseragent5m_window_start" ON "silver"."httptraffic_user_agent_5m" ("window_start");
