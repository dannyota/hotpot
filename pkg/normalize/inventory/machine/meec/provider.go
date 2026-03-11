package meec

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"danny.vn/hotpot/pkg/normalize/inventory/machine"
)

const (
	key         = "meec"
	label       = "MEEC"
	bronzeTable = "meec_inventory_computers"
)

// Provider normalizes bronze.meec_inventory_computers into NormalizedMachine records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return false }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]machine.NormalizedMachine, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, resource_name,
			CASE os_platform WHEN 1 THEN 'windows' WHEN 2 THEN 'macos' WHEN 3 THEN 'linux' ELSE 'unknown' END,
			COALESCE(os_name, ''),
			computer_live_status,
			collected_at, first_collected_at,
			COALESCE(mac_address, ''),
			agent_installed_on
		FROM bronze.meec_inventory_computers`)
	if err != nil {
		return nil, fmt.Errorf("query meec computers: %w", err)
	}
	defer rows.Close()

	var result []machine.NormalizedMachine
	for rows.Next() {
		var resourceID, hostname, osType, osName, macRaw string
		var liveStatus int
		var collectedAt, firstCollectedAt sql.NullTime
		var agentInstalledOn sql.NullInt64
		if err := rows.Scan(&resourceID, &hostname, &osType, &osName,
			&liveStatus, &collectedAt, &firstCollectedAt, &macRaw, &agentInstalledOn); err != nil {
			return nil, fmt.Errorf("scan meec computer: %w", err)
		}

		var macs []string
		for _, mac := range strings.Split(macRaw, ",") {
			if norm := machine.NormalizeMAC(strings.TrimSpace(mac)); norm != "" {
				macs = append(macs, norm)
			}
		}

		status := "running"
		if liveStatus != 1 {
			status = "stopped"
		}

		var created *time.Time
		if agentInstalledOn.Valid && agentInstalledOn.Int64 > 0 {
			t := time.UnixMilli(agentInstalledOn.Int64)
			created = &t
		}
		result = append(result, machine.NormalizedMachine{
			Provider:         key,
			IsBase:           false,
			BronzeTable:      bronzeTable,
			BronzeResourceID: resourceID,
			Hostname:         hostname,
			OSType:           osType,
			OSName:           osName,
			Status:           status,
			Created:          created,
			CollectedAt:      collectedAt.Time,
			FirstCollectedAt: firstCollectedAt.Time,
			MergeKeys: map[string][]string{
				"mac": macs,
			},
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meec computers: %w", err)
	}
	return result, nil
}
