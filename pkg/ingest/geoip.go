package ingest

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/base/config"
	"danny.vn/hotpot/pkg/base/geoip"
)

// GeoIPActivities holds dependencies for GeoIP download activities.
type GeoIPActivities struct {
	configService *config.Service
}

// UpdateGeoIPResult holds the result of a GeoIP file update.
type UpdateGeoIPResult struct {
	Updated int
	Skipped int
}

// updateGeoIPFilesActivity is the activity function reference for Temporal registration.
var updateGeoIPFilesActivity = (*GeoIPActivities).UpdateGeoIPFiles

// UpdateGeoIPFiles checks and downloads city (DB-IP) + ASN (IPinfo) .mmdb files.
// Uses If-Modified-Since to skip re-downloading unchanged files.
func (a *GeoIPActivities) UpdateGeoIPFiles(ctx context.Context) (*UpdateGeoIPResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking GeoIP files for updates")

	client := geoip.NewClient()
	var updated, skipped int

	// City: DB-IP (gzipped).
	cityPath := a.configService.GeoIPCityPath()
	if cityPath == "" {
		skipped++
	} else {
		url := geoip.CityDownloadURL()
		logger.Info("Checking city GeoIP file", "url", url)
		ok, err := client.DownloadGzipped(ctx, url, cityPath)
		if err != nil {
			return nil, fmt.Errorf("download city: %w", err)
		}
		if ok {
			logger.Info("City GeoIP file updated", "path", cityPath)
			updated++
		} else {
			logger.Info("City GeoIP file is current", "path", cityPath)
			skipped++
		}
	}

	// ASN: IPinfo (raw) if token configured, otherwise DB-IP (gzipped).
	// If the source changed (e.g. token added/removed), delete the old file
	// to force a fresh download from the new source.
	asnPath := a.configService.GeoIPASNPath()
	token := a.configService.Config().AccessLog.IPInfoToken
	if asnPath == "" {
		skipped++
	} else if token != "" {
		if !geoip.ASNSourceIsIPInfo(asnPath) {
			// Source changed from DB-IP to IPinfo — remove old file.
			os.Remove(asnPath)
		}
		url := geoip.ASNDownloadURL(token)
		logger.Info("Checking IPinfo ASN file")
		ok, err := client.DownloadRaw(ctx, url, asnPath)
		if err != nil {
			return nil, fmt.Errorf("download asn (ipinfo): %w", err)
		}
		if ok {
			logger.Info("IPinfo ASN file updated", "path", asnPath)
			updated++
		} else {
			logger.Info("IPinfo ASN file is current", "path", asnPath)
			skipped++
		}
	} else {
		if geoip.ASNSourceIsIPInfo(asnPath) {
			// Source changed from IPinfo to DB-IP — remove old file.
			os.Remove(asnPath)
		}
		url := geoip.ASNDbipDownloadURL()
		logger.Info("Checking DB-IP ASN file (no IPinfo token configured)", "url", url)
		ok, err := client.DownloadGzipped(ctx, url, asnPath)
		if err != nil {
			return nil, fmt.Errorf("download asn (dbip): %w", err)
		}
		if ok {
			logger.Info("DB-IP ASN file updated", "path", asnPath)
			updated++
		} else {
			logger.Info("DB-IP ASN file is current", "path", asnPath)
			skipped++
		}
	}

	logger.Info("GeoIP update complete", "updated", updated, "skipped", skipped)
	return &UpdateGeoIPResult{Updated: updated, Skipped: skipped}, nil
}

// UpdateGeoIPWorkflow downloads updated GeoIP .mmdb files from DB-IP and IPinfo.
func UpdateGeoIPWorkflow(ctx workflow.Context) (*UpdateGeoIPResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting UpdateGeoIPWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    30 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var result UpdateGeoIPResult
	if err := workflow.ExecuteActivity(activityCtx, updateGeoIPFilesActivity).
		Get(ctx, &result); err != nil {
		logger.Error("UpdateGeoIPFiles failed", "error", err)
		return nil, err
	}

	logger.Info("UpdateGeoIPWorkflow complete",
		"updated", result.Updated, "skipped", result.Skipped)
	return &result, nil
}
