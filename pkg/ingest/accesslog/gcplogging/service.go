package gcplogging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	entaccesslog "danny.vn/hotpot/pkg/storage/ent/accesslog"
	"danny.vn/hotpot/pkg/storage/ent/accesslog/bronzeaccesslogingestcursor"
)

// Service handles the ingestion business logic for BigQuery Log Analytics traffic.
type Service struct {
	bqClient  *BQClient
	entClient *entaccesslog.Client
}

// NewService creates a new BigQuery Log Analytics traffic service.
func NewService(bqClient *BQClient, entClient *entaccesslog.Client) *Service {
	return &Service{
		bqClient:  bqClient,
		entClient: entClient,
	}
}

// IngestParams holds parameters for traffic ingestion.
type IngestParams struct {
	Name            string
	SourceType      string
	Role            string
	BigQueryTable   string
	BQFilter        string
	FieldMapping    map[string]string
	IntervalMinutes int

	// Backfill settings for first run (no cursor).
	BackfillDays            int
	BackfillIntervalMinutes int
}

// IngestResult holds the result of traffic ingestion.
type IngestResult struct {
	Name            string
	WindowsIngested int
	CountsCreated   int
}

// sourceKey builds a deterministic hash from source-specific config fields.
func sourceKey(params IngestParams) string {
	raw := params.Name + ":" + params.BigQueryTable
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:8])
}

// Ingest queries BigQuery Log Analytics and stores aggregated counts.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	interval := time.Duration(params.IntervalMinutes) * time.Minute
	if interval <= 0 {
		interval = 5 * time.Minute
	}

	sKey := sourceKey(params)

	// Look up cursor by unique (name, source_type, source_key).
	cursor, err := s.entClient.BronzeAccesslogIngestCursor.Query().
		Where(
			bronzeaccesslogingestcursor.NameEQ(params.Name),
			bronzeaccesslogingestcursor.SourceTypeEQ(params.SourceType),
			bronzeaccesslogingestcursor.SourceKeyEQ(sKey),
		).
		Only(ctx)
	if err != nil && !entaccesslog.IsNotFound(err) {
		return nil, fmt.Errorf("read cursor for %s: %w", params.Name, err)
	}

	// Determine start time.
	var startTime time.Time
	if cursor != nil {
		// Update role if it changed.
		if cursor.Role != params.Role {
			_, err = s.entClient.BronzeAccesslogIngestCursor.UpdateOne(cursor).
				SetRole(params.Role).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update cursor role for %s: %w", params.Name, err)
			}
		}
		startTime = cursor.LastWindowEnd
	} else {
		if params.BackfillDays > 0 {
			// First run with backfill — use larger windows.
			backfillInterval := time.Duration(params.BackfillIntervalMinutes) * time.Minute
			if backfillInterval <= 0 {
				backfillInterval = time.Hour
			}
			interval = backfillInterval
			startTime = time.Now().Add(-time.Duration(params.BackfillDays) * 24 * time.Hour).Truncate(interval)
		} else {
			// First run without backfill — 1 hour ago.
			startTime = time.Now().Add(-1 * time.Hour).Truncate(interval)
		}
	}

	// Determine end time: latest complete window before now.
	now := time.Now()
	endTime := now.Truncate(interval)
	if endTime.After(now) {
		endTime = endTime.Add(-interval)
	}

	// Log backfill estimate on first run.
	if cursor == nil && params.BackfillDays > 0 && startTime.Before(endTime) {
		totalWindows := int(endTime.Sub(startTime) / interval)
		slog.InfoContext(ctx, "First run: starting backfill",
			"name", params.Name,
			"backfillDays", params.BackfillDays,
			"intervalMinutes", int(interval.Minutes()),
			"windows", totalWindows,
		)
	}

	if !startTime.Before(endTime) {
		return &IngestResult{Name: params.Name}, nil
	}

	result := &IngestResult{Name: params.Name}
	totalWindows := int(endTime.Sub(startTime) / interval)
	ingestStart := time.Now()

	// Process each window.
	for windowStart := startTime; windowStart.Before(endTime); windowStart = windowStart.Add(interval) {
		windowEnd := windowStart.Add(interval)

		// Query 1: HTTP counts.
		httpCounts, err := s.bqClient.QueryHttpCounts(ctx, params.FieldMapping, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query http counts for window %s: %w", windowStart, err)
		}

		// Query 2: User agents.
		userAgents, err := s.bqClient.QueryUserAgents(ctx, params.FieldMapping, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query user agents for window %s: %w", windowStart, err)
		}

		// Query 3: Client IPs.
		clientIPs, err := s.bqClient.QueryClientIPs(ctx, params.FieldMapping, windowStart, windowEnd)
		if err != nil {
			return nil, fmt.Errorf("query client ips for window %s: %w", windowStart, err)
		}

		result.WindowsIngested++
		if result.WindowsIngested%10 == 0 || result.WindowsIngested == totalWindows {
			elapsed := time.Since(ingestStart)
			rateStr := "N/A"
			etaStr := "N/A"
			if elapsed.Seconds() > 0 {
				rate := float64(result.WindowsIngested) / elapsed.Seconds() * 60
				rateStr = fmt.Sprintf("%.0f/min", rate)
				if rate > 0 {
					remaining := totalWindows - result.WindowsIngested
					eta := time.Duration(float64(remaining)/rate*60) * time.Second
					etaStr = eta.Round(time.Second).String()
				}
			}
			slog.InfoContext(ctx, "Ingest progress",
				"name", params.Name,
				"window", fmt.Sprintf("%d/%d", result.WindowsIngested, totalWindows),
				"httpCounts", len(httpCounts),
				"userAgents", len(userAgents),
				"clientIPs", len(clientIPs),
				"elapsed", elapsed.Round(time.Second),
				"rate", rateStr,
				"eta", etaStr,
			)
		}

		// Store HTTP counts.
		collectedAt := time.Now()
		for _, row := range httpCounts {
			resourceID := fmt.Sprintf("%s:%s:%s:%s:%d",
				params.Name,
				windowStart.Format(time.RFC3339),
				row.URI, row.Method, row.StatusCode)

			create := s.entClient.BronzeAccesslogHttpCount.Create().
				SetID(resourceID).
				SetSourceID(params.Name).
				SetWindowStart(windowStart).
				SetWindowEnd(windowEnd).
				SetURI(row.URI).
				SetStatusCode(int(row.StatusCode)).
				SetRequestCount(row.RequestCount).
				SetTotalBodyBytesSent(row.TotalBodyBytes).
				SetCollectedAt(collectedAt).
				SetFirstCollectedAt(collectedAt)

			if row.Method != "" {
				create.SetMethod(row.Method)
			}
			if row.HTTPHost != "" {
				create.SetHTTPHost(row.HTTPHost)
			}
			if row.TotalRequestTime > 0 {
				create.SetTotalRequestTime(row.TotalRequestTime)
			}
			if row.MaxRequestTime > 0 {
				create.SetMaxRequestTime(row.MaxRequestTime)
			}

			if err := create.Exec(ctx); err != nil {
				if !entaccesslog.IsConstraintError(err) {
					return nil, fmt.Errorf("create http count %s: %w", resourceID, err)
				}
				continue
			}
			result.CountsCreated++
		}

		// Store user agents.
		for _, row := range userAgents {
			uaHash := sha256Short(row.UserAgent)
			uaResourceID := fmt.Sprintf("%s:%s:%s:%s:%s",
				params.Name, windowStart.Format(time.RFC3339),
				row.URI, row.Method, uaHash)

			uaCreate := s.entClient.BronzeAccesslogUserAgent.Create().
				SetID(uaResourceID).
				SetSourceID(params.Name).
				SetWindowStart(windowStart).
				SetWindowEnd(windowEnd).
				SetURI(row.URI).
				SetMethod(row.Method).
				SetUserAgent(row.UserAgent).
				SetRequestCount(row.RequestCount).
				SetCollectedAt(collectedAt).
				SetFirstCollectedAt(collectedAt)

			if err := uaCreate.Exec(ctx); err != nil {
				if !entaccesslog.IsConstraintError(err) {
					return nil, fmt.Errorf("create user agent %s: %w", uaResourceID, err)
				}
			}
		}

		// Store client IPs.
		for _, row := range clientIPs {
			ipResourceID := fmt.Sprintf("%s:%s:%s:%s:%s",
				params.Name, windowStart.Format(time.RFC3339),
				row.URI, row.Method, row.ClientIP)

			ipCreate := s.entClient.BronzeAccesslogClientIp.Create().
				SetID(ipResourceID).
				SetSourceID(params.Name).
				SetWindowStart(windowStart).
				SetWindowEnd(windowEnd).
				SetURI(row.URI).
				SetMethod(row.Method).
				SetClientIP(row.ClientIP).
				SetRequestCount(row.RequestCount).
				SetCollectedAt(collectedAt).
				SetFirstCollectedAt(collectedAt)

			if err := ipCreate.Exec(ctx); err != nil {
				if !entaccesslog.IsConstraintError(err) {
					return nil, fmt.Errorf("create client ip %s: %w", ipResourceID, err)
				}
			}
		}

		// Upsert cursor after each window.
		collectedAt = time.Now()
		if cursor != nil {
			_, err = s.entClient.BronzeAccesslogIngestCursor.UpdateOne(cursor).
				SetLastWindowEnd(windowEnd).
				SetCollectedAt(collectedAt).
				Save(ctx)
		} else {
			cursor, err = s.entClient.BronzeAccesslogIngestCursor.Create().
				SetName(params.Name).
				SetSourceType(params.SourceType).
				SetSourceKey(sKey).
				SetRole(params.Role).
				SetLastWindowEnd(windowEnd).
				SetCollectedAt(collectedAt).
				SetFirstCollectedAt(collectedAt).
				Save(ctx)
		}
		if err != nil {
			return nil, fmt.Errorf("update cursor for %s: %w", params.Name, err)
		}
	}

	return result, nil
}

// sha256Short returns the first 8 hex chars of the SHA-256 hash of s.
func sha256Short(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:4])
}
