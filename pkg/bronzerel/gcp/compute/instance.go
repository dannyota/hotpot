// Package compute provides bronze relationship queries for GCP Compute resources.
package compute

import "danny.vn/hotpot/pkg/bronzerel"

// InstanceFirewalls returns firewalls applying to a GCP compute instance.
// Matches on same network + target tags (or no target tags = applies to all).
func InstanceFirewalls() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_firewalls",
		From: `SELECT DISTINCT f.* FROM "bronze"."gcp_compute_firewalls" f
			WHERE f.network IN (
				SELECT n.network FROM "bronze"."gcp_compute_instance_nics" n
				WHERE n."bronze_gcp_compute_instance_nics" = $1
			)
			AND (
				f.target_tags_json IS NULL
				OR f.target_tags_json::text = '[]'
				OR f.target_tags_json::text = 'null'
				OR EXISTS (
					SELECT 1 FROM jsonb_array_elements_text(f.target_tags_json::jsonb) tag
					JOIN "bronze"."gcp_compute_instance_tags" t ON t.tag = tag
					WHERE t."bronze_gcp_compute_instance_tags" = $1
				)
			)`,
	}
}

// InstanceGroups returns instance groups containing a GCP compute instance.
func InstanceGroups() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_instance_groups",
		From: `WITH inst AS (
				SELECT self_link FROM "bronze"."gcp_compute_instances" WHERE resource_id = $1
			)
			SELECT g.* FROM "bronze"."gcp_compute_instance_groups" g
			WHERE g.resource_id IN (
				SELECT m."bronze_gcp_compute_instance_group_members"
				FROM "bronze"."gcp_compute_instance_group_members" m
				WHERE m.instance_url = (SELECT self_link FROM inst)
			)`,
	}
}

// InstanceForwardingRules returns forwarding rules reaching a GCP compute instance.
// Covers all paths:
//  1. target → target_instances.instance = self_link
//  2. target → target_pools.instances_json contains self_link
//  3. backend_service → backend_service_backends.group → instance_group → members
//  4. backend_service → backend_service_backends.group → NEG → neg_endpoints.instance
func InstanceForwardingRules() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_forwarding_rules",
		From: `WITH inst AS (
				SELECT self_link, name FROM "bronze"."gcp_compute_instances" WHERE resource_id = $1
			)
			SELECT DISTINCT fr.* FROM "bronze"."gcp_compute_forwarding_rules" fr
			WHERE fr.target IN (
				SELECT ti.self_link FROM "bronze"."gcp_compute_target_instances" ti
				WHERE ti.instance = (SELECT self_link FROM inst)
			)
			OR fr.target IN (
				SELECT tp.self_link FROM "bronze"."gcp_compute_target_pools" tp
				WHERE tp.instances_json IS NOT NULL
				AND tp.instances_json::jsonb @> to_jsonb(ARRAY[(SELECT self_link FROM inst)])
			)
			OR fr.backend_service IN (
				SELECT bs.self_link FROM "bronze"."gcp_compute_backend_services" bs
				WHERE bs.resource_id IN (
					SELECT b."bronze_gcp_compute_backend_service_backends"
					FROM "bronze"."gcp_compute_backend_service_backends" b
					WHERE b."group" IN (
						SELECT g.self_link FROM "bronze"."gcp_compute_instance_groups" g
						WHERE g.resource_id IN (
							SELECT m."bronze_gcp_compute_instance_group_members"
							FROM "bronze"."gcp_compute_instance_group_members" m
							WHERE m.instance_url = (SELECT self_link FROM inst)
						)
					)
				)
			)
			OR fr.backend_service IN (
				SELECT bs.self_link FROM "bronze"."gcp_compute_backend_services" bs
				WHERE bs.resource_id IN (
					SELECT b."bronze_gcp_compute_backend_service_backends"
					FROM "bronze"."gcp_compute_backend_service_backends" b
					WHERE b."group" IN (
						SELECT n.self_link FROM "bronze"."gcp_compute_negs" n
						WHERE n.name IN (
							SELECT ne.neg_name FROM "bronze"."gcp_compute_neg_endpoints" ne
							WHERE ne.instance = (SELECT name FROM inst)
						)
					)
				)
			)`,
	}
}

// InstanceAddresses returns reserved/static IP addresses assigned to a GCP compute instance.
// Ephemeral IPs are on NICs, not in the addresses table.
func InstanceAddresses() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_addresses",
		From: `SELECT a.* FROM "bronze"."gcp_compute_addresses" a
			WHERE a.users_json IS NOT NULL
			AND a.users_json::jsonb @> to_jsonb(ARRAY[(
				SELECT self_link FROM "bronze"."gcp_compute_instances" WHERE resource_id = $1
			)])`,
	}
}

// InstanceImages returns images related to a GCP compute instance's disks.
// Both directions: images used to create the VM's disks, and images created from them.
func InstanceImages() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_images",
		From: `SELECT DISTINCT i.* FROM "bronze"."gcp_compute_images" i
			WHERE i.self_link IN (
				SELECT d.source_image FROM "bronze"."gcp_compute_disks" d
				WHERE d.self_link IN (
					SELECT ad.source FROM "bronze"."gcp_compute_instance_disks" ad
					WHERE ad."bronze_gcp_compute_instance_disks" = $1 AND ad.source IS NOT NULL AND ad.source != ''
				) AND d.source_image IS NOT NULL AND d.source_image != ''
			)
			OR i.source_disk IN (
				SELECT ad.source FROM "bronze"."gcp_compute_instance_disks" ad
				WHERE ad."bronze_gcp_compute_instance_disks" = $1 AND ad.source IS NOT NULL AND ad.source != ''
			)`,
	}
}

// InstanceSnapshots returns snapshots related to a GCP compute instance's disks.
// Both directions: snapshots of the VM's disks, and snapshots used to create them.
func InstanceSnapshots() bronzerel.Relation {
	return bronzerel.Relation{
		Schema: "bronze",
		Table:  "gcp_compute_snapshots",
		From: `SELECT DISTINCT s.* FROM "bronze"."gcp_compute_snapshots" s
			WHERE s.source_disk IN (
				SELECT ad.source FROM "bronze"."gcp_compute_instance_disks" ad
				WHERE ad."bronze_gcp_compute_instance_disks" = $1 AND ad.source IS NOT NULL AND ad.source != ''
			)
			OR s.self_link IN (
				SELECT d.source_snapshot FROM "bronze"."gcp_compute_disks" d
				WHERE d.self_link IN (
					SELECT ad.source FROM "bronze"."gcp_compute_instance_disks" ad
					WHERE ad."bronze_gcp_compute_instance_disks" = $1 AND ad.source IS NOT NULL AND ad.source != ''
				) AND d.source_snapshot IS NOT NULL AND d.source_snapshot != ''
			)`,
	}
}
