package topic

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TopicDiff represents changes between old and new topic state.
type TopicDiff struct {
	IsNew     bool
	IsChanged bool
}

// HasAnyChange returns true if there are any changes.
func (d *TopicDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

// DiffTopicData compares existing Ent entity with new TopicData and returns differences.
func DiffTopicData(old *ent.BronzeGCPPubSubTopic, new *TopicData) *TopicDiff {
	diff := &TopicDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.KmsKeyName != new.KmsKeyName ||
		old.MessageRetentionDuration != new.MessageRetentionDuration ||
		old.State != new.State ||
		!bytes.Equal(old.LabelsJSON, new.LabelsJSON) ||
		!bytes.Equal(old.MessageStoragePolicyJSON, new.MessageStoragePolicyJSON) ||
		!bytes.Equal(old.SchemaSettingsJSON, new.SchemaSettingsJSON) ||
		!bytes.Equal(old.IngestionDataSourceSettingsJSON, new.IngestionDataSourceSettingsJSON) {
		diff.IsChanged = true
	}

	return diff
}
