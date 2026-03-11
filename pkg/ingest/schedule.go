package ingest

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	hotpottemporal "danny.vn/hotpot/pkg/base/temporal"
	"danny.vn/hotpot/pkg/base/config"
)

// ensureSchedules creates paused daily schedules for all enabled providers.
// Existing schedules are left unchanged.
func ensureSchedules(ctx context.Context, temporalClient client.Client, providers []ProviderRegistration, configService *config.Service) {
	sc := temporalClient.ScheduleClient()

	for _, p := range providers {
		if !p.Enabled(configService) || p.Workflow == nil {
			continue
		}

		hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
			ID: fmt.Sprintf("hotpot-ingest-%s-daily", p.Name),
			Spec: client.ScheduleSpec{
				Intervals: []client.ScheduleIntervalSpec{
					{Every: 24 * time.Hour},
				},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        fmt.Sprintf("hotpot-ingest-%s", p.Name),
				Workflow:  p.Workflow,
				Args:      p.WorkflowArgs,
				TaskQueue: p.TaskQueue,
			},
			Paused: true,
		})
	}

	// GeoIP download schedule — runs on its own task queue, unpaused.
	hotpottemporal.EnsureSchedule(ctx, sc, client.ScheduleOptions{
		ID: "hotpot-ingest-geoip-daily",
		Spec: client.ScheduleSpec{
			Intervals: []client.ScheduleIntervalSpec{
				{Every: 24 * time.Hour},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			ID:        "hotpot-ingest-geoip",
			Workflow:  UpdateGeoIPWorkflow,
			TaskQueue: "hotpot-ingest-geoip",
		},
		Paused: false,
	})

	// Trigger immediate GeoIP download if files don't exist.
	triggerGeoIPDownloadIfNeeded(ctx, temporalClient, configService)
}


// triggerGeoIPDownloadIfNeeded starts the GeoIP download workflow if the
// default mmdb files are missing and no run is already in progress.
func triggerGeoIPDownloadIfNeeded(ctx context.Context, temporalClient client.Client, configService *config.Service) {
	cityPath := configService.GeoIPCityPath()
	asnPath := configService.GeoIPASNPath()

	cityExists := fileExists(cityPath)
	asnExists := fileExists(asnPath)
	if cityExists && asnExists {
		return
	}

	log.Printf("GeoIP files missing (city=%s exists=%v, asn=%s exists=%v), triggering download",
		cityPath, cityExists, asnPath, asnExists)

	workflowID := "hotpot-ingest-geoip"
	_, err := temporalClient.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:                    workflowID,
		TaskQueue:             "hotpot-ingest-geoip",
		WorkflowIDReusePolicy:    enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
	}, UpdateGeoIPWorkflow)
	if err != nil {
		log.Printf("Failed to trigger GeoIP download: %v", err)
		return
	}

	log.Printf("Triggered GeoIP download workflow: %s", workflowID)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
