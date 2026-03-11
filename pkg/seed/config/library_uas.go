package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedLibraryUAs inserts predefined HTTP library user-agent families.
func SeedLibraryUAs(ctx context.Context, db *sql.DB) error {
	if len(libraryUAs) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.library_uas
		(family, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(libraryUAs)*5)
	for i, e := range libraryUAs {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5)
		args = append(args, e.family, e.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (family) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert library uas (%d entries): %w", len(libraryUAs), err)
	}
	return nil
}

type libraryUAEntry struct {
	family      string
	description string
}

var libraryUAs = []libraryUAEntry{
	{"curl", "cURL command-line HTTP client"},
	{"python-requests", "Python requests library"},
	{"go-http-client", "Go standard HTTP client"},
	{"wget", "GNU Wget"},
	{"httpie", "HTTPie CLI"},
	{"java", "Java HTTP client"},
	{"okhttp", "OkHttp (Java/Kotlin)"},
	{"axios", "Axios (Node.js)"},
	{"node-fetch", "Node.js fetch"},
	{"php-curl", "PHP cURL"},
}
