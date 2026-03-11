package accesslog

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"danny.vn/hotpot/pkg/base/config"
	entaccesslog "danny.vn/hotpot/pkg/storage/ent/accesslog"
	"danny.vn/hotpot/pkg/storage/ent/accesslog/bronzeaccesslogclientip"
	"danny.vn/hotpot/pkg/storage/ent/accesslog/bronzeaccessloghttpcount"
	"danny.vn/hotpot/pkg/storage/ent/accesslog/bronzeaccessloguseragent"
)

// Activities holds dependencies for accesslog provider-level Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entaccesslog.Client
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entaccesslog.Client) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
	}
}

// Activity function references for Temporal registration.
var (
	DiscoverLogSourcesActivity = (*Activities).DiscoverLogSources
	CleanupStaleBronzeActivity = (*Activities).CleanupStaleBronze
)

// DiscoverLogSourcesResult holds the discovered active log sources.
type DiscoverLogSourcesResult struct {
	Sources []LogSourceInfo
}

// LogSourceInfo is a serializable summary of a log source from config.
type LogSourceInfo struct {
	Name            string
	SourceType      string
	Role            string
	ProjectID       string
	BigQueryTable   string
	BQFilter        string
	FieldMapping    map[string]string
	IntervalMinutes int

	// Backfill settings (global, applied to all sources).
	BackfillDays            int
	BackfillIntervalMinutes int
}

// DiscoverLogSources reads log sources from config.
func (a *Activities) DiscoverLogSources(ctx context.Context) (*DiscoverLogSourcesResult, error) {
	logger := activity.GetLogger(ctx)

	sources := a.configService.AccessLogSources()
	backfillDays := a.configService.AccessLogBackfillDays()
	backfillInterval := a.configService.AccessLogBackfillIntervalMinutes()

	result := &DiscoverLogSourcesResult{
		Sources: make([]LogSourceInfo, 0, len(sources)),
	}
	for _, s := range sources {
		result.Sources = append(result.Sources, LogSourceInfo{
			Name:                    s.Name,
			SourceType:              s.Type,
			Role:                    s.Role,
			ProjectID:               s.ProjectID,
			BigQueryTable:           s.BigQueryTable,
			BQFilter:                s.BQFilter,
			FieldMapping:            s.FieldMapping,
			IntervalMinutes:         s.IntervalMinutes,
			BackfillDays:            backfillDays,
			BackfillIntervalMinutes: backfillInterval,
		})
	}

	logger.Info("Discovered log sources from config", "count", len(result.Sources))
	return result, nil
}

// CleanupStaleBronzeResult holds cleanup statistics.
type CleanupStaleBronzeResult struct {
	BronzeCountsDeleted     int
	BronzeUserAgentsDeleted int
	BronzeClientIPsDeleted  int
}

// CleanupStaleBronze removes old bronze traffic data beyond retention.
func (a *Activities) CleanupStaleBronze(ctx context.Context) (*CleanupStaleBronzeResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Cleaning up stale bronze traffic data")

	days := a.configService.AccessLogRetentionDays()
	if days <= 0 {
		logger.Info("Retention days not configured, skipping cleanup")
		return &CleanupStaleBronzeResult{}, nil
	}
	cutoff := time.Now().Add(-time.Duration(days) * 24 * time.Hour)

	deleted, err := a.entClient.BronzeAccesslogHttpCount.Delete().
		Where(bronzeaccessloghttpcount.WindowStartLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete old bronze counts: %w", err)
	}

	deletedUA, err := a.entClient.BronzeAccesslogUserAgent.Delete().
		Where(bronzeaccessloguseragent.WindowStartLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete old bronze user agents: %w", err)
	}

	deletedIP, err := a.entClient.BronzeAccesslogClientIp.Delete().
		Where(bronzeaccesslogclientip.WindowStartLT(cutoff)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete old bronze client IPs: %w", err)
	}

	logger.Info("Bronze cleanup complete",
		"bronzeCountsDeleted", deleted,
		"bronzeUserAgentsDeleted", deletedUA,
		"bronzeClientIPsDeleted", deletedIP)
	return &CleanupStaleBronzeResult{
		BronzeCountsDeleted:     deleted,
		BronzeUserAgentsDeleted: deletedUA,
		BronzeClientIPsDeleted:  deletedIP,
	}, nil
}
