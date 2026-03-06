package s1

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"danny.vn/hotpot/pkg/normalize/inventory/software"
)

const (
	key         = "s1"
	label       = "SentinelOne"
	bronzeTable = "s1_endpoint_apps"
)

// Provider normalizes bronze.s1_endpoint_apps into NormalizedSoftware records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return true }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]software.NormalizedSoftware, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT ea.resource_id, m.resource_id,
			LOWER(ea.name), COALESCE(ea.version, ''), COALESCE(ea.publisher, ''),
			ea.collected_at, ea.first_collected_at
		FROM bronze.s1_endpoint_apps ea
		JOIN bronze.s1_agents a ON ea.agent_id = a.resource_id
		JOIN inventory.machine_bronze_links l
			ON l.bronze_resource_id = a.resource_id AND l.bronze_table = 's1_agents'
		JOIN inventory.machines m ON m.resource_id = l.inventory_machine_bronze_links
		WHERE ea.name IS NOT NULL AND ea.name != ''`)
	if err != nil {
		return nil, fmt.Errorf("query s1 endpoint apps: %w", err)
	}
	defer rows.Close()

	var result []software.NormalizedSoftware
	for rows.Next() {
		var bronzeID, machineID, name, version, publisher string
		var collectedAt, firstCollectedAt sql.NullTime
		if err := rows.Scan(&bronzeID, &machineID, &name, &version, &publisher,
			&collectedAt, &firstCollectedAt); err != nil {
			return nil, fmt.Errorf("scan s1 endpoint app: %w", err)
		}
		result = append(result, software.NormalizedSoftware{
			Provider:         key,
			IsBase:           true,
			BronzeTable:      bronzeTable,
			BronzeResourceID: bronzeID,
			MachineID:        machineID,
			Name:             strings.TrimSpace(name),
			Version:          strings.TrimSpace(version),
			Publisher:        strings.TrimSpace(publisher),
			CollectedAt:      collectedAt.Time,
			FirstCollectedAt: firstCollectedAt.Time,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate s1 endpoint apps: %w", err)
	}
	return result, nil
}
