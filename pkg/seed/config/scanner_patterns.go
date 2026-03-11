package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedScannerPatterns inserts predefined security scanner UA keywords.
func SeedScannerPatterns(ctx context.Context, db *sql.DB) error {
	if len(scannerPatterns) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.scanner_patterns
		(keyword, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(scannerPatterns)*5)
	for i, e := range scannerPatterns {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5)
		args = append(args, e.keyword, e.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (keyword) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert scanner patterns (%d entries): %w", len(scannerPatterns), err)
	}
	return nil
}

type scannerEntry struct {
	keyword     string
	description string
}

var scannerPatterns = []scannerEntry{
	{"sqlmap", "SQL injection tool"},
	{"nikto", "Web server scanner"},
	{"nmap", "Network scanner"},
	{"masscan", "Mass IP port scanner"},
	{"zgrab", "Application layer scanner"},
	{"nuclei", "Vulnerability scanner"},
	{"gobuster", "Directory/file brute-forcer"},
	{"dirbuster", "Directory brute-forcer"},
	{"wfuzz", "Web fuzzer"},
}
