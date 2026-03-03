package greennode

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dannyota/hotpot/pkg/normalize/machine"
)

const (
	key         = "greennode"
	label       = "GreenNode"
	bronzeTable = "greennode_compute_servers"
)

// Provider normalizes bronze.greennode_compute_servers into NormalizedMachine records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]machine.NormalizedMachine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, name, status, region,
			COALESCE(flavor_name, ''),
			COALESCE(image_version, ''),
			COALESCE(server_group_name, ''),
			collected_at, first_collected_at,
			COALESCE(interfaces_json::text, '{}')
		FROM bronze.greennode_compute_servers`)
	if err != nil {
		return nil, fmt.Errorf("query greennode servers: %w", err)
	}
	defer rows.Close()

	var result []machine.NormalizedMachine
	for rows.Next() {
		var resourceID, hostname, status, region, flavorName, imageVer, serverGroup, ifacesJSON string
		var collectedAt, firstCollectedAt sql.NullTime
		if err := rows.Scan(&resourceID, &hostname, &status, &region,
			&flavorName, &imageVer, &serverGroup, &collectedAt, &firstCollectedAt, &ifacesJSON); err != nil {
			return nil, fmt.Errorf("scan greennode server: %w", err)
		}

		// STOPPED = deleted in GreenNode.
		if status == "STOPPED" {
			continue
		}

		macs, firstIP := parseInterfaces(ifacesJSON)

		result = append(result, machine.NormalizedMachine{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: resourceID,
			Hostname:         hostname,
			OSType:           imageToOSType(imageVer),
			OSName:           imageVer,
			Status:           normalizeStatus(status),
			InternalIP:       firstIP,
			Environment:      machine.InferEnvironment(hostname, ""),
			CloudProject:     serverGroup,
			CloudZone:        region,
			CloudMachineType: flavorName,
			CollectedAt:      collectedAt.Time,
			FirstCollectedAt: firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"mac": macs,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate greennode servers: %w", err)
	}
	return result, nil
}

func parseInterfaces(jsonStr string) (macs []string, firstIP string) {
	var ifaces map[string][]struct {
		FixedIP string `json:"fixedIp"`
		MAC     string `json:"mac"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &ifaces); err != nil {
		return nil, ""
	}
	for _, nics := range ifaces {
		for _, nic := range nics {
			if norm := machine.NormalizeMAC(nic.MAC); norm != "" {
				macs = append(macs, norm)
			}
			if firstIP == "" && strings.TrimSpace(nic.FixedIP) != "" {
				firstIP = strings.TrimSpace(nic.FixedIP)
			}
		}
	}
	return macs, firstIP
}

func imageToOSType(imageVersion string) string {
	lower := strings.ToLower(imageVersion)
	switch {
	case strings.Contains(lower, "windows"):
		return "windows"
	case strings.Contains(lower, "redhat"), strings.Contains(lower, "centos"),
		strings.Contains(lower, "ubuntu"), strings.Contains(lower, "linux"),
		strings.Contains(lower, "debian"), strings.Contains(lower, "rocky"):
		return "linux"
	default:
		return "unknown"
	}
}

func normalizeStatus(s string) string {
	switch s {
	case "ACTIVE":
		return "running"
	default:
		return strings.ToLower(s)
	}
}
