package httptraffic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	enthttptraffic "danny.vn/hotpot/pkg/storage/ent/httptraffic"
)

// Activity function reference for Temporal registration.
var NormalizeUserAgentsActivity = (*Activities).NormalizeUserAgents

// NormalizeUserAgentsResult holds normalization statistics.
type NormalizeUserAgentsResult struct {
	Processed int
	Mapped    int
}

// NormalizeUserAgents reads bronze user agent data, enriches with endpoint match + UA parsing, writes to silver.
func (a *Activities) NormalizeUserAgents(ctx context.Context, params NormalizeTrafficParams) (*NormalizeUserAgentsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Normalizing user agents")

	since := params.Since

	// Load endpoints and build path matcher.
	pm, err := a.getPathMatcher(ctx)
	if err != nil {
		return nil, err
	}

	// Query bronze user agents.
	rows, err := a.db.QueryContext(ctx, `
		SELECT source_id, window_start, window_end, uri,
			COALESCE(method, ''), user_agent, request_count, collected_at
		FROM bronze.accesslog_user_agents
		WHERE collected_at >= $1
		ORDER BY window_start`, since)
	if err != nil {
		return nil, fmt.Errorf("query bronze user agents: %w", err)
	}
	defer rows.Close()

	now := time.Now()
	var processed, mapped int

	for rows.Next() {
		var sourceID, uri, method, userAgent string
		var windowStart, windowEnd, collectedAt time.Time
		var requestCount int64
		if err := rows.Scan(&sourceID, &windowStart, &windowEnd, &uri,
			&method, &userAgent, &requestCount, &collectedAt); err != nil {
			return nil, fmt.Errorf("scan bronze user agent: %w", err)
		}

		ep := pm.Match(uri)
		uaFamily := ParseUAFamily(userAgent)
		uaHash := sha256Short(userAgent)
		resourceID := fmt.Sprintf("%s:%s:%s:%s:%s",
			sourceID, windowStart.Format(time.RFC3339),
			uri, method, uaHash)

		create := a.entClient.SilverHttptrafficUserAgent5m.Create().
			SetID(resourceID).
			SetSourceID(sourceID).
			SetWindowStart(windowStart).
			SetWindowEnd(windowEnd).
			SetURI(uri).
			SetUserAgent(userAgent).
			SetUaFamily(uaFamily).
			SetRequestCount(requestCount).
			SetCollectedAt(collectedAt).
			SetFirstCollectedAt(collectedAt).
			SetNormalizedAt(now)
		if method != "" {
			create.SetMethod(method)
		}
		if ep != nil {
			create.SetEndpointID(ep.ID).SetIsMapped(true)
			mapped++
		}
		processed++

		if err := create.Exec(ctx); err != nil {
			if !enthttptraffic.IsConstraintError(err) {
				return nil, fmt.Errorf("create user_agent_5m %s: %w", resourceID, err)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bronze user agents: %w", err)
	}

	logger.Info("User agent normalization complete",
		"processed", processed, "mapped", mapped)
	return &NormalizeUserAgentsResult{Processed: processed, Mapped: mapped}, nil
}

func sha256Short(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:4])
}
