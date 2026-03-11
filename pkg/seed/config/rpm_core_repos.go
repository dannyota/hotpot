package config

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SeedRpmCoreRepos inserts predefined RPM core repository definitions.
func SeedRpmCoreRepos(ctx context.Context, db *sql.DB) error {
	if len(rpmCoreRepos) == 0 {
		return nil
	}

	now := time.Now()
	var b strings.Builder
	b.WriteString(`INSERT INTO config.rpm_core_repos
		(repo_name, description, is_active, created_at, updated_at)
		VALUES `)

	args := make([]any, 0, len(rpmCoreRepos)*5)
	for i, e := range rpmCoreRepos {
		if i > 0 {
			b.WriteString(", ")
		}
		base := i * 5
		fmt.Fprintf(&b, "($%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5)
		args = append(args, e.repoName, e.description, true, now, now)
	}

	b.WriteString(` ON CONFLICT (repo_name) DO NOTHING`)

	_, err := db.ExecContext(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("upsert rpm core repos (%d entries): %w", len(rpmCoreRepos), err)
	}
	return nil
}

type rpmRepoEntry struct {
	repoName    string
	description string
}

var rpmCoreRepos = []rpmRepoEntry{
	{"epel7", "EPEL 7 repository"},
	{"epel9", "EPEL 9 repository"},
	{"rhel7-extras", "RHEL 7 Extras"},
	{"rhel7-os", "RHEL 7 Base OS"},
	{"rhel7-sclo", "RHEL 7 Software Collections"},
	{"rhel7-updates", "RHEL 7 Updates"},
	{"rhel9-appstream", "RHEL 9 AppStream"},
	{"rhel9-baseos", "RHEL 9 BaseOS"},
	{"rhel9-crb", "RHEL 9 CodeReady Builder"},
	{"rhel9-ha", "RHEL 9 High Availability"},
}
