package notificationconfig

import (
	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// NotificationConfigDiff represents changes between old and new SCC notification config state.
type NotificationConfigDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *NotificationConfigDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffNotificationConfigData compares existing Ent entity with new NotificationConfigData and returns differences.
func DiffNotificationConfigData(old *ent.BronzeGCPSecurityCenterNotificationConfig, new *NotificationConfigData) *NotificationConfigDiff {
	diff := &NotificationConfigDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.PubsubTopic != new.PubsubTopic ||
		old.StreamingConfigJSON != new.StreamingConfigJSON ||
		old.ServiceAccount != new.ServiceAccount {
		diff.IsChanged = true
	}

	return diff
}
