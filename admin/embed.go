package admin

import "embed"

// DistFS holds the built Vue frontend.
// Embedded directly from admin/ui/dist/ — no copy step needed.
//
//go:embed all:ui/dist
var DistFS embed.FS
