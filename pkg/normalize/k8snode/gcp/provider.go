package gcp

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dannyota/hotpot/pkg/normalize/k8snode"
)

const (
	key         = "gcp"
	label       = "GCP GKE"
	bronzeTable = "gcp_compute_instances"
)

// Provider normalizes bronze.gcp_compute_instances GKE nodes into NormalizedK8sNode records.
// Only instances with the goog-gke-node label are included.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]k8snode.NormalizedK8sNode, error) {
	// Load GKE labels into maps keyed by resource_id.
	type labelSet struct {
		ClusterName  string
		NodePool     string
		Provisioning string
	}
	labels := make(map[string]*labelSet)

	labelRows, err := db.QueryContext(ctx, `
		SELECT bronze_gcp_compute_instance_labels, key, value
		FROM bronze.gcp_compute_instance_labels
		WHERE key IN ('goog-gke-node', 'goog-k8s-cluster-name', 'goog-k8s-node-pool-name', 'goog-gke-node-pool-provisioning-model')`)
	if err != nil {
		return nil, fmt.Errorf("query gke labels: %w", err)
	}
	defer labelRows.Close()

	for labelRows.Next() {
		var resourceID, k, value string
		if err := labelRows.Scan(&resourceID, &k, &value); err != nil {
			return nil, fmt.Errorf("scan gke label: %w", err)
		}
		ls, ok := labels[resourceID]
		if !ok {
			ls = &labelSet{}
			labels[resourceID] = ls
		}
		switch k {
		case "goog-k8s-cluster-name":
			ls.ClusterName = value
		case "goog-k8s-node-pool-name":
			ls.NodePool = value
		case "goog-gke-node-pool-provisioning-model":
			ls.Provisioning = value
		}
	}
	if err := labelRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gke labels: %w", err)
	}

	if len(labels) == 0 {
		return nil, nil
	}

	// Load instances + NICs for GKE nodes only.
	rows, err := db.QueryContext(ctx, `
		SELECT i.resource_id, i.name,
			COALESCE(i.zone, ''), COALESCE(i.machine_type, ''),
			COALESCE(i.status, ''), i.project_id,
			i.collected_at, i.first_collected_at,
			COALESCE(n.network_ip, ''),
			COALESCE(ac.nat_ip, '')
		FROM bronze.gcp_compute_instances i
		JOIN bronze.gcp_compute_instance_labels l
			ON l.bronze_gcp_compute_instance_labels = i.resource_id AND l.key = 'goog-gke-node'
		LEFT JOIN bronze.gcp_compute_instance_nics n
			ON n.bronze_gcp_compute_instance_nics = i.resource_id
		LEFT JOIN bronze.gcp_compute_instance_nic_access_configs ac
			ON ac.bronze_gcp_compute_instance_nic_access_configs = n.id`)
	if err != nil {
		return nil, fmt.Errorf("query gke node instances: %w", err)
	}
	defer rows.Close()

	type nodeEntry struct {
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
	}

	seen := make(map[string]*nodeEntry)
	var order []string

	for rows.Next() {
		var resourceID, name, zone, machineType, status, projectID, networkIP, natIP string
		var collectedAt, firstCollectedAt sql.NullTime
		if err := rows.Scan(&resourceID, &name, &zone, &machineType, &status, &projectID,
			&collectedAt, &firstCollectedAt, &networkIP, &natIP); err != nil {
			return nil, fmt.Errorf("scan gke node instance: %w", err)
		}
		if existing, ok := seen[resourceID]; ok {
			if existing.internalIP == "" && networkIP != "" {
				existing.internalIP = networkIP
			}
			if existing.externalIP == "" && natIP != "" {
				existing.externalIP = natIP
			}
			continue
		}

		seen[resourceID] = &nodeEntry{
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
		order = append(order, resourceID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gke node instances: %w", err)
	}

	result := make([]k8snode.NormalizedK8sNode, 0, len(order))
	for _, id := range order {
		inst := seen[id]
		ls := labels[id]
		if ls == nil {
			ls = &labelSet{}
		}

		result = append(result, k8snode.NormalizedK8sNode{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: inst.resourceID,
			NodeName:         inst.name,
			ClusterName:      ls.ClusterName,
			NodePool:         ls.NodePool,
			Status:           normalizeStatus(inst.status),
			Provisioning:     ls.Provisioning,
			CloudProject:     inst.projectID,
			CloudZone:        shortName(inst.zone),
			CloudMachineType: shortName(inst.machineType),
			InternalIP:       inst.internalIP,
			ExternalIP:       inst.externalIP,
			CollectedAt:      inst.collectedAt.Time,
			FirstCollectedAt: inst.firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"internal_ip": {inst.internalIP},
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
