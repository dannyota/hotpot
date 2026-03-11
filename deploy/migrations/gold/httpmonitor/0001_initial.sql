-- Add new schema named "gold"
CREATE SCHEMA IF NOT EXISTS "gold";
-- Create "httpmonitor_anomalies" table
CREATE TABLE "gold"."httpmonitor_anomalies" (
  "resource_id" character varying NOT NULL,
  "detected_at" timestamptz NOT NULL,
  "first_detected_at" timestamptz NOT NULL,
  "endpoint_id" character varying NULL,
  "source_id" character varying NOT NULL,
  "anomaly_type" character varying NOT NULL,
  "severity" character varying NOT NULL,
  "window_start" timestamptz NOT NULL,
  "window_end" timestamptz NOT NULL,
  "uri" character varying NULL,
  "method" character varying NULL,
  "baseline_value" double precision NULL,
  "actual_value" double precision NULL,
  "deviation" double precision NULL,
  "description" character varying NULL,
  "evidence_json" jsonb NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "goldhttpmonitoranomaly_anomaly_type" to table: "httpmonitor_anomalies"
CREATE INDEX "goldhttpmonitoranomaly_anomaly_type" ON "gold"."httpmonitor_anomalies" ("anomaly_type");
-- Create index "goldhttpmonitoranomaly_endpoint_id" to table: "httpmonitor_anomalies"
CREATE INDEX "goldhttpmonitoranomaly_endpoint_id" ON "gold"."httpmonitor_anomalies" ("endpoint_id");
-- Create index "goldhttpmonitoranomaly_severity" to table: "httpmonitor_anomalies"
CREATE INDEX "goldhttpmonitoranomaly_severity" ON "gold"."httpmonitor_anomalies" ("severity");
-- Create index "goldhttpmonitoranomaly_source_id_window_start" to table: "httpmonitor_anomalies"
CREATE INDEX "goldhttpmonitoranomaly_source_id_window_start" ON "gold"."httpmonitor_anomalies" ("source_id", "window_start");
