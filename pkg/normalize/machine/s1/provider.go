package s1

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dannyota/hotpot/pkg/normalize/machine"
)

const (
	key        = "s1"
	label      = "SentinelOne"
	bronzeTable = "s1_agents"
)

// bestOSName picks the most descriptive name from S1's os_name and os_revision.
// S1's os_name is generic for Linux/macOS ("Linux", "macOS") but specific for
// Windows ("Windows 11 Pro"). os_revision has detail for Linux/macOS
// ("Red Hat Enterprise release 9.4 ...") but only build numbers for Windows ("26200").
// For Windows, both are combined: "Windows 11 Pro (Build 26200)".
func bestOSName(osType, osName, osRevision string) string {
	generic := strings.EqualFold(osName, osType) || strings.EqualFold(osName, "linux") || strings.EqualFold(osName, "macos")
	if osName != "" && !generic {
		if osRevision != "" {
			return osName + " (Build " + osRevision + ")"
		}
		return osName
	}
	if osRevision != "" {
		return osRevision
	}
	return osName
}

// Provider normalizes bronze.s1_agents into NormalizedMachine records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]machine.NormalizedMachine, error) {
	// Load agents.
	agentRows, err := db.QueryContext(ctx, `
		SELECT resource_id, computer_name, os_type,
			COALESCE(os_name, ''), COALESCE(os_revision, ''),
			COALESCE(site_name, ''), COALESCE(last_ip_to_mgmt, ''),
			is_active, is_decommissioned, is_uninstalled,
			collected_at, first_collected_at
		FROM bronze.s1_agents`)
	if err != nil {
		return nil, fmt.Errorf("query s1 agents: %w", err)
	}
	defer agentRows.Close()

	type agent struct {
		id               string
		hostname         string
		osType           string
		osName           string
		osRevision       string
		site             string
		lastIPToMgmt     string
		isActive         bool
		isDecommissioned bool
		isUninstalled    bool
		collectedAt      sql.NullTime
		firstCollectedAt sql.NullTime
		macs             []string
		ips              []string
	}

	agents := make(map[string]*agent)
	for agentRows.Next() {
		var a agent
		if err := agentRows.Scan(&a.id, &a.hostname, &a.osType, &a.osName, &a.osRevision,
			&a.site, &a.lastIPToMgmt, &a.isActive, &a.isDecommissioned, &a.isUninstalled,
			&a.collectedAt, &a.firstCollectedAt); err != nil {
			return nil, fmt.Errorf("scan s1 agent: %w", err)
		}
		agents[a.id] = &a
	}
	if err := agentRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate s1 agents: %w", err)
	}

	// Load NICs for merge keys, excluding:
	// 1. Virtual/bridge interfaces by name (ap*, awdl*, bridge*, docker*, veth*, etc.)
	// 2. MACs shared across 3+ agents (dynamically detected — e.g., GCP internal, VMware, macOS Thunderbolt dummy)
	nicRows, err := db.QueryContext(ctx, `
		SELECT n.bronze_s1agent_nics, COALESCE(n.physical, ''),
			COALESCE(n.inet_json::text, '[]')
		FROM bronze.s1_agent_nics n
		JOIN bronze.s1_agents a ON n.bronze_s1agent_nics = a.resource_id
		WHERE LOWER(n.name) NOT LIKE ANY(ARRAY[
			'ap%', 'awdl%', 'llw%', 'p2p%', 'bridge%', 'vmenet%', 'vmnet%', 'utun%',
			'docker%', 'br-%', 'veth%', 'virbr%', 'cni%', 'flannel%', 'calico%', 'tun%', 'tap%',
			'hyper-v%', 'vethernet%', 'lo', 'loopback', 'stf%'
		])
		AND UPPER(n.physical) NOT IN (
			SELECT UPPER(n2.physical) FROM bronze.s1_agent_nics n2
			WHERE n2.physical IS NOT NULL AND n2.physical != ''
			GROUP BY UPPER(n2.physical) HAVING COUNT(DISTINCT n2.bronze_s1agent_nics) >= 3
		)`)
	if err != nil {
		return nil, fmt.Errorf("query s1 agent nics: %w", err)
	}
	defer nicRows.Close()

	for nicRows.Next() {
		var agentID, mac, inetJSON string
		if err := nicRows.Scan(&agentID, &mac, &inetJSON); err != nil {
			return nil, fmt.Errorf("scan s1 agent nic: %w", err)
		}
		a, ok := agents[agentID]
		if !ok {
			continue
		}
		if norm := machine.NormalizeMAC(mac); norm != "" {
			a.macs = append(a.macs, norm)
		}
		var ips []string
		if err := json.Unmarshal([]byte(inetJSON), &ips); err == nil {
			for _, ip := range ips {
				if ip = strings.TrimSpace(ip); ip != "" {
					a.ips = append(a.ips, ip)
				}
			}
		}
	}
	if err := nicRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate s1 agent nics: %w", err)
	}

	// Build normalized records, filtering out deleted agents.
	var result []machine.NormalizedMachine
	for _, a := range agents {
		if a.isDecommissioned || a.isUninstalled {
			continue
		}
		status := "running"
		if !a.isActive {
			status = "stopped"
		}
		internalIP := ""
		if len(a.ips) > 0 {
			internalIP = a.ips[0]
		} else if a.lastIPToMgmt != "" {
			internalIP = a.lastIPToMgmt
			a.ips = append(a.ips, a.lastIPToMgmt)
		}
		result = append(result, machine.NormalizedMachine{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: a.id,
			Hostname:         a.hostname,
			OSType:           a.osType,
			OSName:           bestOSName(a.osType, a.osName, a.osRevision),
			Status:           status,
			InternalIP:       internalIP,
			Environment:      machine.InferEnvironment(a.hostname, a.site),
			CollectedAt:      a.collectedAt.Time,
			FirstCollectedAt: a.firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"mac": a.macs,
			},
		})
	}
	return result, nil
}
