package meec

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dannyota/hotpot/pkg/normalize/installedsoftware"
)

const (
	key         = "meec"
	label       = "MEEC"
	bronzeTable = "meec_inventory_installed_software"
)

// Provider normalizes bronze.meec_inventory_installed_software into NormalizedInstalledSoftware records.
type Provider struct{}

func (Provider) Key() string   { return key }
func (Provider) Label() string { return label }
func (Provider) IsBase() bool  { return false }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]installedsoftware.NormalizedInstalledSoftware, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT is2.resource_id, m.resource_id,
			LOWER(COALESCE(is2.software_name, '')),
			COALESCE(is2.software_version, ''),
			COALESCE(is2.manufacturer_name, ''),
			is2.collected_at, is2.first_collected_at
		FROM bronze.meec_inventory_installed_software is2
		JOIN bronze.meec_inventory_computers c ON is2.computer_resource_id = c.resource_id
		JOIN silver.machine_bronze_links l
			ON l.bronze_resource_id = c.resource_id AND l.bronze_table = 'meec_inventory_computers'
		JOIN silver.machines m ON m.resource_id = l.silver_machine_bronze_links
		WHERE is2.software_name IS NOT NULL AND is2.software_name != ''`)
	if err != nil {
		return nil, fmt.Errorf("query meec installed software: %w", err)
	}
	defer rows.Close()

	var result []installedsoftware.NormalizedInstalledSoftware
	for rows.Next() {
		var bronzeID, machineID, name, version, publisher string
		var collectedAt, firstCollectedAt sql.NullTime
		if err := rows.Scan(&bronzeID, &machineID, &name, &version, &publisher,
			&collectedAt, &firstCollectedAt); err != nil {
			return nil, fmt.Errorf("scan meec installed software: %w", err)
		}
		result = append(result, installedsoftware.NormalizedInstalledSoftware{
			Provider:         key,
			IsBase:           false,
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
		return nil, fmt.Errorf("iterate meec installed software: %w", err)
	}
	return result, nil
}
