package gcp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"danny.vn/hotpot/pkg/normalize/inventory/machine"
)

const (
	key         = "gcp"
	label       = "GCP Compute"
	bronzeTable = "gcp_compute_instances"
)

// Provider normalizes bronze.gcp_compute_instances into NormalizedMachine records.
// GKE node instances (identified by goog-gke-node label) are excluded.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]machine.NormalizedMachine, error) {
	// Load GKE node instance IDs to exclude.
	gkeRows, err := db.QueryContext(ctx, `
		SELECT DISTINCT bronze_gcp_compute_instance_labels
		FROM bronze.gcp_compute_instance_labels
		WHERE key = 'goog-gke-node'`)
	if err != nil {
		return nil, fmt.Errorf("query gke node labels: %w", err)
	}
	defer gkeRows.Close()

	gkeNodes := make(map[string]bool)
	for gkeRows.Next() {
		var resourceID string
		if err := gkeRows.Scan(&resourceID); err != nil {
			return nil, fmt.Errorf("scan gke node label: %w", err)
		}
		gkeNodes[resourceID] = true
	}
	if err := gkeRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gke node labels: %w", err)
	}

	// Load instances with NIC and access config data.
	rows, err := db.QueryContext(ctx, `
		SELECT i.resource_id, i.name,
			COALESCE(i.zone, ''), COALESCE(i.machine_type, ''),
			COALESCE(i.status, ''), i.project_id,
			i.collected_at, i.first_collected_at,
			COALESCE(n.network_ip, ''),
			COALESCE(ac.nat_ip, '')
		FROM bronze.gcp_compute_instances i
		LEFT JOIN bronze.gcp_compute_instance_nics n
			ON n.bronze_gcp_compute_instance_nics = i.resource_id
		LEFT JOIN bronze.gcp_compute_instance_nic_access_configs ac
			ON ac.bronze_gcp_compute_instance_nic_access_configs = n.id`)
	if err != nil {
		return nil, fmt.Errorf("query gcp compute instances: %w", err)
	}
	defer rows.Close()

	type instance struct {
		resourceID       string
		name             string
		zone             string
		machineType      string
		status           string
		projectID        string
		collectedAt      sql.NullTime
		firstCollectedAt sql.NullTime
		internalIP       string
		externalIP       string
		ips              []string
	}

	seen := make(map[string]*instance)
	var order []string

	for rows.Next() {
		var resourceID, name, zone, machineType, status, projectID, networkIP, natIP string
		var collectedAt, firstCollectedAt sql.NullTime
		if err := rows.Scan(&resourceID, &name, &zone, &machineType, &status, &projectID,
			&collectedAt, &firstCollectedAt, &networkIP, &natIP); err != nil {
			return nil, fmt.Errorf("scan gcp compute instance: %w", err)
		}
		if gkeNodes[resourceID] {
			continue
		}
		if existing, ok := seen[resourceID]; ok {
			if existing.internalIP == "" && networkIP != "" {
				existing.internalIP = networkIP
			}
			if existing.externalIP == "" && natIP != "" {
				existing.externalIP = natIP
			}
			if networkIP != "" {
				existing.ips = append(existing.ips, networkIP)
			}
			continue
		}
		inst := &instance{
			resourceID:       resourceID,
			name:             name,
			zone:             zone,
			machineType:      machineType,
			status:           status,
			projectID:        projectID,
			collectedAt:      collectedAt,
			firstCollectedAt: firstCollectedAt,
			internalIP:       networkIP,
			externalIP:       natIP,
		}
		if networkIP != "" {
			inst.ips = append(inst.ips, networkIP)
		}
		seen[resourceID] = inst
		order = append(order, resourceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gcp compute instances: %w", err)
	}

	result := make([]machine.NormalizedMachine, 0, len(order))
	for _, id := range order {
		inst := seen[id]
		status := normalizeStatus(inst.status)
		result = append(result, machine.NormalizedMachine{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: inst.resourceID,
			Hostname:         inst.name,
			OSType:           "linux",
			Status:           status,
			InternalIP:       inst.internalIP,
			ExternalIP:       inst.externalIP,
			CloudProject:     inst.projectID,
			CloudZone:        shortName(inst.zone),
			CloudMachineType: shortName(inst.machineType),
			CollectedAt:      inst.collectedAt.Time,
			FirstCollectedAt: inst.firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"internal_ip": inst.ips,
			},
		})
	}
	return result, nil
}

func normalizeStatus(s string) string {
	switch s {
	case "RUNNING":
		return "running"
	case "TERMINATED":
		return "stopped"
	default:
		return strings.ToLower(s)
	}
}

func shortName(url string) string {
	if i := strings.LastIndex(url, "/"); i >= 0 {
		return url[i+1:]
	}
	return url
}
