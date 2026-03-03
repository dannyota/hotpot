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

// Provider normalizes bronze.s1_agents into NormalizedMachine records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]machine.NormalizedMachine, error) {
	// Load agents.
	agentRows, err := db.QueryContext(ctx, `
		SELECT resource_id, computer_name, os_type,
			COALESCE(os_revision, ''),
			COALESCE(site_name, ''), is_active,
			is_decommissioned, is_uninstalled,
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
		site             string
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
		if err := agentRows.Scan(&a.id, &a.hostname, &a.osType, &a.osName,
			&a.site, &a.isActive, &a.isDecommissioned, &a.isUninstalled,
			&a.collectedAt, &a.firstCollectedAt); err != nil {
			return nil, fmt.Errorf("scan s1 agent: %w", err)
		}
		agents[a.id] = &a
	}
	if err := agentRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate s1 agents: %w", err)
	}

	// Load NICs for merge keys.
	nicRows, err := db.QueryContext(ctx, `
		SELECT n.bronze_s1agent_nics, COALESCE(n.physical, ''),
			COALESCE(n.inet_json::text, '[]')
		FROM bronze.s1_agent_nics n
		JOIN bronze.s1_agents a ON n.bronze_s1agent_nics = a.resource_id`)
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
		}
		result = append(result, machine.NormalizedMachine{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: a.id,
			Hostname:         a.hostname,
			OSType:           a.osType,
			OSName:           a.osName,
			Status:           status,
			InternalIP:       internalIP,
			Environment:      machine.InferEnvironment(a.hostname, a.site),
			CollectedAt:      a.collectedAt.Time,
			FirstCollectedAt: a.firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"mac":         a.macs,
				"internal_ip": a.ips,
			},
		})
	}
	return result, nil
}
