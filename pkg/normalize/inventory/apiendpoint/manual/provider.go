package manual

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"danny.vn/hotpot/pkg/normalize/inventory/apiendpoint"
)

const (
	key         = "manual"
	bronzeTable = "apicatalog_endpoints_raw"
)

// Provider normalizes bronze.apicatalog_endpoints_raw into NormalizedApiEndpoint records.
type Provider struct{}

func (Provider) Key() string { return key }

func (Provider) Load(ctx context.Context, db *sql.DB) ([]apiendpoint.NormalizedApiEndpoint, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT resource_id, uri,
			COALESCE(name, ''),
			COALESCE(upstream, ''),
			method, COALESCE(route_status, ''),
			collected_at, first_collected_at
		FROM bronze.apicatalog_endpoints_raw`)
	if err != nil {
		return nil, fmt.Errorf("query apicatalog_endpoints_raw: %w", err)
	}
	defer rows.Close()

	var result []apiendpoint.NormalizedApiEndpoint
	for rows.Next() {
		var (
			resourceID, uri, name, upstream string
			methodRaw, routeStatus          string
			collectedAt, firstCollectedAt   sql.NullTime
		)
		if err := rows.Scan(&resourceID, &uri, &name, &upstream,
			&methodRaw, &routeStatus,
			&collectedAt, &firstCollectedAt); err != nil {
			return nil, fmt.Errorf("scan apicatalog row: %w", err)
		}

		// Split "POST,PUT" → ["POST", "PUT"].
		var methods []string
		if methodRaw != "" {
			for _, m := range strings.Split(methodRaw, ",") {
				m = strings.TrimSpace(m)
				if m != "" {
					methods = append(methods, strings.ToUpper(m))
				}
			}
		}

		// Derive active from route_status.
		isActive := strings.EqualFold(routeStatus, "active")

		// Derive access_level from URI prefix.
		accessLevel := deriveAccessLevel(uri)

		result = append(result, apiendpoint.NormalizedApiEndpoint{
			BronzeResourceID: resourceID,
			Name:             name,
			Service:          upstream,
			URIPattern:       uri,
			Methods:          methods,
			IsActive:         isActive,
			AccessLevel:      accessLevel,
			Provider:         key,
			BronzeTable:      bronzeTable,
			CollectedAt:      collectedAt.Time,
			FirstCollectedAt: firstCollectedAt.Time,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate apicatalog rows: %w", err)
	}

	return result, nil
}

// deriveAccessLevel infers access level from URI prefix.
func deriveAccessLevel(uri string) string {
	lower := strings.ToLower(uri)
	switch {
	case strings.HasPrefix(lower, "/public/"):
		return "public"
	case strings.HasPrefix(lower, "/protected/"):
		return "protected"
	case strings.HasPrefix(lower, "/private/"):
		return "private"
	default:
		return ""
	}
}
