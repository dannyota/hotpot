package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedSanctionedCountries inserts predefined sanctioned country codes.
func SeedSanctionedCountries(ctx context.Context, db *sql.DB) error {
	if len(sanctionedCountries) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.sanctioned_countries
		(country_code, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(sanctionedCountries)*5)
	for i, e := range sanctionedCountries {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5)
		args = append(args, e.code, e.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (country_code) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert sanctioned countries (%d entries): %w", len(sanctionedCountries), err)
	}
	return nil
}

type sanctionedEntry struct {
	code        string
	description string
}

// OFAC Comprehensive Sanctions Programs as of 2025.
// https://ofac.treasury.gov/sanctions-programs-and-country-information
var sanctionedCountries = []sanctionedEntry{
	{"CU", "Cuba (OFAC)"},
	{"IR", "Iran (OFAC)"},
	{"KP", "North Korea (OFAC)"},
	{"SY", "Syria (OFAC)"},
}
