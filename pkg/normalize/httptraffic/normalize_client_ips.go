package httptraffic

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	enthttptraffic "danny.vn/hotpot/pkg/storage/ent/httptraffic"
)

// Activity function reference for Temporal registration.
var NormalizeClientIPsActivity = (*Activities).NormalizeClientIPs

// NormalizeClientIPsResult holds normalization statistics.
type NormalizeClientIPsResult struct {
	Processed int
	Mapped    int
	GeoHits   int
	ASNHits   int
}

// NormalizeClientIPs reads bronze client IP data, enriches with endpoint match + GeoIP/ASN, writes to silver.
func (a *Activities) NormalizeClientIPs(ctx context.Context, params NormalizeTrafficParams) (*NormalizeClientIPsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing client IPs")

	since := params.Since

	// Load endpoints and build path matcher.
	pm, err := a.getPathMatcher(ctx)
	if err != nil {
		return nil, err
	}

	// Reload GeoIP files to pick up any updates.
	a.geoip.Reload()

	// Query bronze client IPs.
	rows, err := a.db.QueryContext(ctx, `
		SELECT source_id, window_start, window_end, uri,
			COALESCE(method, ''), client_ip, request_count, collected_at
		FROM bronze.accesslog_client_ips
		WHERE collected_at >= $1
		ORDER BY window_start`, since)
	if err != nil {
		return nil, fmt.Errorf("query bronze client IPs: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var processed, mapped, geoHits, asnHits int

	for rows.Next() {
		var sourceID, uri, method, clientIP string
		var windowStart, windowEnd, collectedAt time.Time
		var requestCount int64
		if err := rows.Scan(&sourceID, &windowStart, &windowEnd, &uri,
			&method, &clientIP, &requestCount, &collectedAt); err != nil {
			return nil, fmt.Errorf("scan bronze client IP: %w", err)
		}

		ep := pm.Match(uri)
		geo := a.geoip.LookupIP(clientIP)

		resourceID := fmt.Sprintf("%s:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339),
			uri, method, clientIP)

		create := a.entClient.SilverHttptrafficClientIp5m.Create().
			SetID(resourceID).
			SetSourceID(sourceID).
			SetWindowStart(windowStart).
			SetWindowEnd(windowEnd).
			SetURI(uri).
			SetClientIP(clientIP).
			SetIsInternal(geo.IsInternal).
			SetRequestCount(requestCount).
			SetCollectedAt(collectedAt).
			SetFirstCollectedAt(collectedAt).
			SetNormalizedAt(now)
		if method != "" {
			create.SetMethod(method)
		}
		if geo.CountryCode != "" {
			create.SetCountryCode(geo.CountryCode).SetCountryName(geo.CountryName)
			geoHits++
		}
		if geo.ASN > 0 {
			create.SetAsn(geo.ASN).SetOrgName(geo.OrgName)
			if geo.ASDomain != "" {
				create.SetAsDomain(geo.ASDomain)
			}
			if geo.ASNType != "" {
				create.SetAsnType(geo.ASNType)
			}
			asnHits++
		}
		if ep != nil {
			create.SetEndpointID(ep.ID).SetIsMapped(true)
			mapped++
		}
		processed++

		if err := create.Exec(ctx); err != nil {
			if !enthttptraffic.IsConstraintError(err) {
				return nil, fmt.Errorf("create client_ip_5m %s: %w", resourceID, err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bronze client IPs: %w", err)
	}

	logger.Info("Client IP normalization complete",
		"processed", processed, "mapped", mapped,
		"geoHits", geoHits, "asnHits", asnHits)
	return &NormalizeClientIPsResult{
		Processed: processed, Mapped: mapped,
		GeoHits: geoHits, ASNHits: asnHits,
	}, nil
}
