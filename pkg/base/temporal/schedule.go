package temporal

import (
	"context"
	"log/slog"

	"go.temporal.io/sdk/client"
)

// EnsureSchedule creates a schedule if it doesn't exist, or updates the action
// (task queue, workflow) and spec if the schedule already exists. The schedule's
// state (paused/unpaused) is preserved on updates.
func EnsureSchedule(ctx context.Context, sc client.ScheduleClient, opts client.ScheduleOptions) {
	handle := sc.GetHandle(ctx, opts.ID)
	if _, err := handle.Describe(ctx); err != nil {
		// Schedule doesn't exist — create it.
		_, err := sc.Create(ctx, opts)
		if err != nil {
			slog.Error("Failed to create schedule", "schedule", opts.ID, "error", err)
			return
		}
		slog.Info("Created schedule", "schedule", opts.ID)
		return
	}

	// Schedule exists — update action and spec, preserving state (paused, etc.).
	err := handle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			s := input.Description.Schedule
			s.Action = opts.Action
			spec := opts.Spec
			s.Spec = &spec
			return &client.ScheduleUpdate{
				Schedule: &s,
			}, nil
		},
	})
	if err != nil {
		slog.Error("Failed to update schedule", "schedule", opts.ID, "error", err)
		return
	}
	slog.Info("Ensured schedule", "schedule", opts.ID)
}
