-- Create "reference_os_core_rules" table
CREATE TABLE "bronze"."reference_os_core_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "rule_type" character varying NOT NULL,
  "os_type" character varying NULL,
  "value" character varying NOT NULL,
  "description" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferenceoscorerule_collected_at" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_collected_at" ON "bronze"."reference_os_core_rules" ("collected_at");
-- Create index "bronzereferenceoscorerule_os_type" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_os_type" ON "bronze"."reference_os_core_rules" ("os_type");
-- Create index "bronzereferenceoscorerule_rule_type" to table: "reference_os_core_rules"
CREATE INDEX "bronzereferenceoscorerule_rule_type" ON "bronze"."reference_os_core_rules" ("rule_type");
-- Create "reference_software_match_rules" table
CREATE TABLE "bronze"."reference_software_match_rules" (
  "resource_id" character varying NOT NULL,
  "collected_at" timestamptz NOT NULL,
  "first_collected_at" timestamptz NOT NULL,
  "product_slug" character varying NOT NULL,
  "rule_type" character varying NOT NULL,
  "os_type" character varying NULL,
  "value" character varying NOT NULL,
  "extra_value" character varying NULL,
  PRIMARY KEY ("resource_id")
);
-- Create index "bronzereferencesoftwarematchrule_collected_at" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_collected_at" ON "bronze"."reference_software_match_rules" ("collected_at");
-- Create index "bronzereferencesoftwarematchrule_product_slug" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_product_slug" ON "bronze"."reference_software_match_rules" ("product_slug");
-- Create index "bronzereferencesoftwarematchrule_rule_type" to table: "reference_software_match_rules"
CREATE INDEX "bronzereferencesoftwarematchrule_rule_type" ON "bronze"."reference_software_match_rules" ("rule_type");
