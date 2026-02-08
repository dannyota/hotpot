// Package migrations embeds the SQL migration files so the production migrate
// binary is self-contained (no extra files needed at deploy time).
package migrations

import "embed"

//go:embed */*.sql */atlas.sum
var FS embed.FS
