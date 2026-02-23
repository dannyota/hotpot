package ingest

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/dannyota/hotpot/pkg/base/config"
)

// ensureSchedules creates paused daily schedules for all enabled providers.
// Existing schedules are left unchanged.
func ensureSchedules(ctx context.Context, temporalClient client.Client, providers []ProviderRegistration, configService *config.Service) {
	sc := temporalClient.ScheduleClient()

	for _, p := range providers {
		if !p.Enabled(configService) || p.Workflow == nil {
			continue
		}

		scheduleID := fmt.Sprintf("hotpot-ingest-%s-daily", p.Name)

		// Skip if schedule already exists.
		handle := sc.GetHandle(ctx, scheduleID)
		if _, err := handle.Describe(ctx); err == nil {
			log.Printf("Schedule %s already exists", scheduleID)
			continue
		}

		_, err := sc.Create(ctx, client.ScheduleOptions{
			ID: scheduleID,
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
		if err != nil {
			log.Printf("Failed to create schedule %s: %v", scheduleID, err)
			continue
		}

		log.Printf("Created paused schedule: %s", scheduleID)
	}
}
