// Package bronzerel provides shared relationship queries between bronze resources.
// Both admin (detail pages) and normalize (silver mapping) use these to traverse
// the GCP resource graph stored in the bronze schema.
//
// Each relationship function returns a Relation whose From field is a SQL clause
// with $1 bound to the parent resource_id. Consumers wrap it with their own
// SELECT/ORDER/LIMIT.
package bronzerel

// Relation describes a query from a parent resource to related resources.
type Relation struct {
	Schema string // PG schema, e.g. "bronze"
	Table  string // target table, e.g. "gcp_compute_firewalls"
	// From is a complete SQL SELECT with $1 = parent resource_id.
	// Consumers wrap as subquery: SELECT ... FROM ({From}) t [WHERE ...] [ORDER BY ...] [LIMIT ...]
	From string
}
