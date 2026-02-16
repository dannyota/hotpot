package subscription

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SubscriptionDiff represents changes between old and new subscription state.
type SubscriptionDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *SubscriptionDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffSubscriptionData compares existing Ent entity with new SubscriptionData and returns differences.
func DiffSubscriptionData(old *ent.BronzeGCPPubSubSubscription, new *SubscriptionData) *SubscriptionDiff {
	diff := &SubscriptionDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Topic != new.Topic ||
		old.AckDeadlineSeconds != new.AckDeadlineSeconds ||
		old.RetainAckedMessages != new.RetainAckedMessages ||
		old.MessageRetentionDuration != new.MessageRetentionDuration ||
		old.EnableMessageOrdering != new.EnableMessageOrdering ||
		old.Filter != new.Filter ||
		old.Detached != new.Detached ||
		old.EnableExactlyOnceDelivery != new.EnableExactlyOnceDelivery ||
		old.State != new.State ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.PushConfigJSON, new.PushConfigJSON) ||
		!bytes.Equal(old.BigqueryConfigJSON, new.BigqueryConfigJSON) ||
		!bytes.Equal(old.CloudStorageConfigJSON, new.CloudStorageConfigJSON) ||
		!bytes.Equal(old.ExpirationPolicyJSON, new.ExpirationPolicyJSON) ||
		!bytes.Equal(old.DeadLetterPolicyJSON, new.DeadLetterPolicyJSON) ||
		!bytes.Equal(old.RetryPolicyJSON, new.RetryPolicyJSON) {
		diff.IsChanged = true
	}

	return diff
}
