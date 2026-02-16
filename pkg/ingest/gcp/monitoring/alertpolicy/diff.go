package alertpolicy

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AlertPolicyDiff represents changes between old and new alert policy state.
type AlertPolicyDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *AlertPolicyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffAlertPolicyData compares existing Ent entity with new AlertPolicyData and returns differences.
func DiffAlertPolicyData(old *ent.BronzeGCPMonitoringAlertPolicy, new *AlertPolicyData) *AlertPolicyDiff {
	diff := &AlertPolicyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.DisplayName != new.DisplayName ||
		old.Combiner != new.Combiner ||
		old.Enabled != new.Enabled ||
		old.Severity != new.Severity ||
		!bytes.Equal(old.DocumentationJSON, new.DocumentationJSON) ||
		!bytes.Equal(old.UserLabelsJSON, new.UserLabelsJSON) ||
		!bytes.Equal(old.ConditionsJSON, new.ConditionsJSON) ||
		!bytes.Equal(old.NotificationChannelsJSON, new.NotificationChannelsJSON) ||
		!bytes.Equal(old.CreationRecordJSON, new.CreationRecordJSON) ||
		!bytes.Equal(old.MutationRecordJSON, new.MutationRecordJSON) ||
		!bytes.Equal(old.AlertStrategyJSON, new.AlertStrategyJSON) {
		diff.IsChanged = true
	}

	return diff
}
