// Package seed provides default data for config tables. Each seeder inserts
// predefined rows using ON CONFLICT DO NOTHING (additive only — user
// modifications are preserved). Called by cmd/migrate after DDL migrations.
package seed

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"danny.vn/hotpot/pkg/seed/config"
)

// Run executes all seeders in order. Tables must already exist (DDL first).
func Run(ctx context.Context, db *sql.DB) error {
	seeders := []struct {
		name string
		fn   func(context.Context, *sql.DB) error
	}{
		// Config tables — detection.
		{"hosting_indicators", config.SeedHostingIndicators},
		{"scanner_patterns", config.SeedScannerPatterns},
		{"library_uas", config.SeedLibraryUAs},
		{"httpmonitor_rules", config.SeedHttpmonitorRules},
		{"sanctioned_countries", config.SeedSanctionedCountries},
		{"uri_attack_patterns", config.SeedURIAttackPatterns},
		{"auth_endpoint_patterns", config.SeedAuthEndpointPatterns},
		// Config tables — lifecycle.
		{"software_match_rules", config.SeedSoftwareMatchRules},
		{"os_core_rules", config.SeedOSCoreRules},
		{"rpm_core_repos", config.SeedRpmCoreRepos},
	}

	for _, s := range seeders {
		log.Printf("  seeding %s...", s.name)
		if err := s.fn(ctx, db); err != nil {
			return fmt.Errorf("seed %s: %w", s.name, err)
		}
	}
	return nil
}
