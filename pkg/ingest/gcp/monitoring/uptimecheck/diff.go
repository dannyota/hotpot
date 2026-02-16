package uptimecheck

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// UptimeCheckDiff represents changes between old and new uptime check config state.
type UptimeCheckDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *UptimeCheckDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffUptimeCheckData compares existing Ent entity with new UptimeCheckData and returns differences.
func DiffUptimeCheckData(old *ent.BronzeGCPMonitoringUptimeCheckConfig, new *UptimeCheckData) *UptimeCheckDiff {
	diff := &UptimeCheckDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.Period != new.Period ||
		old.Timeout != new.Timeout ||
		old.CheckerType != new.CheckerType ||
		old.IsInternal != new.IsInternal ||
		!bytes.Equal(old.MonitoredResourceJSON, new.MonitoredResourceJSON) ||
		!bytes.Equal(old.ResourceGroupJSON, new.ResourceGroupJSON) ||
		!bytes.Equal(old.HTTPCheckJSON, new.HttpCheckJSON) ||
		!bytes.Equal(old.TCPCheckJSON, new.TcpCheckJSON) ||
		!bytes.Equal(old.ContentMatchersJSON, new.ContentMatchersJSON) ||
		!bytes.Equal(old.SelectedRegionsJSON, new.SelectedRegionsJSON) ||
		!bytes.Equal(old.InternalCheckersJSON, new.InternalCheckersJSON) ||
		!bytes.Equal(old.UserLabelsJSON, new.UserLabelsJSON) {
		diff.IsChanged = true
	}

	return diff
}
